package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const schedulerName = "metricscheduler"

type predicateFunc func(node *v1.Node, pod *v1.Pod) bool
type priorityFunc func(node *v1.Node, pod *v1.Pod) int

type SchedulerParams struct {
	SchedulerName string
	Timeout       int
	LogLevel      string
	FilteredNodes string
}

type Scheduler struct {
	clientset       *kubernetes.Clientset
	podQueue        chan *v1.Pod
	schedulerParams SchedulerParams
	nodeLister      listersv1.NodeLister
	predicates      []predicateFunc
	priorities      []priorityFunc
}

func NewScheduler(podQueue chan *v1.Pod, quit chan struct{}) Scheduler {

	log.SetFormatter(&log.JSONFormatter{})
	var params = SchedulerParams{}

	flag.StringVar(&params.SchedulerName, "scheduler-name", LookupEnvOrString("SCHEDULER_NAME", "random"), "scheduler name.")
	flag.StringVar(&params.LogLevel, "log-level", LookupEnvOrString("LOG_LEVEL", "info"), "scheduler log level.")
	flag.StringVar(&params.FilteredNodes, "filtered-nodes", LookupEnvOrString("FILTERED_NODES", ""), "Nodes to filer.")
	flag.IntVar(&params.Timeout, "timeout", LookupEnvOrInt("TIMEOUT", 20), "Timeout connecting in seconds")

	flag.Parse()

	switch params.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	log.SetOutput(os.Stdout)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return Scheduler{
		schedulerParams: params,
		clientset:       clientset,
		podQueue:        podQueue,
		nodeLister:      initInformers(clientset, podQueue, quit, params.SchedulerName),
		predicates: []predicateFunc{
			randomPredicate,
		},
		priorities: []priorityFunc{
			randomPriority,
		},
	}
}

func initInformers(clientset *kubernetes.Clientset, podQueue chan *v1.Pod, quit chan struct{}, schedulerName string) listersv1.NodeLister {
	factory := informers.NewSharedInformerFactory(clientset, 0)

	nodeInformer := factory.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*v1.Node)
			if !ok {
				log.Println("this is not a node")
				return
			}
			log.Printf("New Node Added to Store: %s", node.GetName())
		},
	})

	podInformer := factory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				log.Println("this is not a pod")
				return
			}
			if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == schedulerName {
				podQueue <- pod
			}
		},
	})

	factory.Start(quit)
	return nodeInformer.Lister()
}

func main() {
	fmt.Println("I'm a scheduler!")

	rand.New(rand.NewSource(time.Now().UnixNano()))

	podQueue := make(chan *v1.Pod, 300)
	defer close(podQueue)

	quit := make(chan struct{})
	defer close(quit)

	scheduler := NewScheduler(podQueue, quit)
	scheduler.Run(quit)
}

func (s *Scheduler) Run(quit chan struct{}) {
	wait.Until(s.ScheduleOne, 0, quit)
}

func (s *Scheduler) ScheduleOne() {
	ctx := context.TODO()

	p := <-s.podQueue
	fmt.Println("found a pod to schedule:", p.Namespace, "/", p.Name)

	node, err := s.findFit(p)
	if err != nil {
		log.Println("cannot find node that fits pod", err.Error())
		return
	}

	err = s.bindPod(ctx, p, node)
	if err != nil {
		log.Println("failed to bind pod", err.Error())
		return
	}

	message := fmt.Sprintf("Placed pod [%s/%s] on %s\n", p.Namespace, p.Name, node)

	err = s.emitEvent(ctx, p, message)
	if err != nil {
		log.Println("failed to emit scheduled event", err.Error())
		return
	}

	fmt.Println(message)
}

func (s *Scheduler) getNodesToInspect(nodes []*v1.Node, userFilteredNodes []string) []*v1.Node {
	filteredNodes := make([]*v1.Node, 0)

	for _, node := range nodes {
		var filter = false
		for _, userNodes := range userFilteredNodes {
			if cmp.Equal(node.Name, userNodes) {
				filter = true
			}
		}
		if !filter {
			filteredNodes = append(filteredNodes, node)
		}
	}

	log.Println("nodes to inspect: ")
	for _, n := range filteredNodes {
		log.Println(n.Name)
	}
	return filteredNodes
}

func (s *Scheduler) findFit(pod *v1.Pod) (string, error) {
	nodes, err := s.nodeLister.List(labels.Everything())
	if err != nil {
		return "", err
	}

	var nodesToInspect []*v1.Node

	if s.schedulerParams.FilteredNodes != "" {
		filteredNodesSlice := strings.Split(s.schedulerParams.FilteredNodes, ",")
		nodesToInspect = s.getNodesToInspect(nodes, filteredNodesSlice)
	} else {
		nodesToInspect = nodes
	}

	//filteredNodes := s.runPredicates(nodesToInspect, pod)
	if len(nodesToInspect) == 0 {
		return "", errors.New("failed to find node that fits pod")
	}
	priorities := s.prioritize(nodesToInspect, pod)
	return s.findBestNode(priorities), nil
}

func (s *Scheduler) bindPod(ctx context.Context, p *v1.Pod, node string) error {
	opts := metav1.CreateOptions{}
	return s.clientset.CoreV1().Pods(p.Namespace).Bind(ctx, &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.Name,
			Namespace: p.Namespace,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node,
		},
	}, opts)
}

func (s *Scheduler) emitEvent(ctx context.Context, p *v1.Pod, message string) error {
	timestamp := time.Now().UTC()
	opts := metav1.CreateOptions{}
	_, err := s.clientset.CoreV1().Events(p.Namespace).Create(ctx, &v1.Event{
		Count:          1,
		Message:        message,
		Reason:         "Scheduled",
		LastTimestamp:  metav1.NewTime(timestamp),
		FirstTimestamp: metav1.NewTime(timestamp),
		Type:           "Normal",
		Source: v1.EventSource{
			//Component: os.Getenv("SCHEDULER_NAME"),
			Component: s.schedulerParams.SchedulerName,
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Pod",
			Name:      p.Name,
			Namespace: p.Namespace,
			UID:       p.UID,
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: p.Name + "-",
		},
	}, opts)
	if err != nil {
		return err
	}
	return nil
}

func (s *Scheduler) runPredicates(nodes []*v1.Node, pod *v1.Pod) []*v1.Node {
	filteredNodes := make([]*v1.Node, 0)
	for _, node := range nodes {
		if s.predicatesApply(node, pod) {
			filteredNodes = append(filteredNodes, node)
		}
	}
	log.Println("nodes that fit:")
	for _, n := range filteredNodes {
		log.Println(n.Name)
	}
	return filteredNodes
}

func (s *Scheduler) predicatesApply(node *v1.Node, pod *v1.Pod) bool {
	for _, predicate := range s.predicates {
		if !predicate(node, pod) {
			return false
		}
	}
	return true
}

func randomPredicate(node *v1.Node, pod *v1.Pod) bool {
	r := rand.Intn(2)
	return r == 0
}

func (s *Scheduler) prioritize(nodes []*v1.Node, pod *v1.Pod) map[string]int {
	priorities := make(map[string]int)
	for _, node := range nodes {
		for _, priority := range s.priorities {
			priorities[node.Name] += priority(node, pod)
		}
	}
	log.Println("calculated priorities:", priorities)
	return priorities
}

func (s *Scheduler) findBestNode(priorities map[string]int) string {
	var maxP int
	var bestNode string
	for node, p := range priorities {
		if p > maxP {
			maxP = p
			bestNode = node
		}
	}
	return bestNode
}

func randomPriority(node *v1.Node, pod *v1.Pod) int {
	return rand.Intn(100)
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Error(err, "LookupEnvOrInt", "key", key, "value", val)
		}
		return v
	}
	return defaultVal
}
