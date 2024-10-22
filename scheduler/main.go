package main

import (
	"flag"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"main/pkg"
	"main/pkg/postgresql"
	"math/rand"
	"os"

	"time"

	"context"
	"errors"
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
)

type predicateFunc func(node *v1.Node, pod *v1.Pod) bool
type priorityFunc func(node *v1.Node, pod *v1.Pod) int

type Scheduler struct {
	schedulerParams pkg.SchedulerParams
	dbClient        postgresql.DatabaseClient

	clientset  *kubernetes.Clientset
	podQueue   chan *v1.Pod
	nodeLister listersv1.NodeLister

	predicates []predicateFunc
}

func NewScheduler(podQueue chan *v1.Pod, quit chan struct{}) Scheduler {

	log.SetFormatter(&log.JSONFormatter{})
	var params = pkg.SchedulerParams{}

	//metrics params
	flag.StringVar(&params.MetricParams.MetricName, "metric-name", "metric-name", "Metric name in Prometheus to scheduled")
	flag.StringVar(&params.MetricParams.StartDate, "metric-start-date", "", "Start date to get metrics")
	flag.StringVar(&params.MetricParams.EndDate, "metric-end-date", "", "End date to get metrics")
	flag.StringVar(&params.MetricParams.Operation, "metric-operation", "", "Operation to get  metrics, example: max,min,avg,...")
	flag.StringVar(&params.MetricParams.PriorityOrder, "metric-priority-order", "", "how to priority results, example. order asc o desc")
	flag.StringVar(&params.MetricParams.FilterClause, "metric-filter-clause", "", "Extra filter clause")
	flag.StringVar(&params.MetricParams.IsSecondLevel, "metric-is-second-level", "", "Is second level")
	flag.StringVar(&params.MetricParams.SecondLevelGroup, "metric-second-level-group", "", "Second level group")
	flag.StringVar(&params.MetricParams.SecondLevelSelect, "metric-second-level-select", "", "Second level select")

	//TimescaleDbParams
	flag.StringVar(&params.TimescaleDbParams.Host, "timescaledb-host", "timescaledb.monitoring", "host to connect to timescaledb")
	flag.StringVar(&params.TimescaleDbParams.Port, "timescaledb-port", "5231", "port to connect to timescaledb")
	flag.StringVar(&params.TimescaleDbParams.User, "timescaledb-user", "postgres", "user to connect to timescaledb")
	flag.StringVar(&params.TimescaleDbParams.Password, "timescaledb-password", "patroni", "password to connect to timescaledb")
	flag.StringVar(&params.TimescaleDbParams.Database, "timescaledb-database", "tsdb", "database name to connect to timescaledb")
	flag.StringVar(&params.TimescaleDbParams.AuthenticationType, "timescaledb-auth-type", "MD5", "database name to connect to timescaledb")

	flag.StringVar(&params.SchedulerName, "scheduler-name", "random", "scheduler name.")
	flag.StringVar(&params.LogLevel, "log-level", "info", "scheduler log level.")
	flag.StringVar(&params.FilteredNodes, "filtered-nodes", "", "Nodes to filer.")
	flag.IntVar(&params.Timeout, "timeout", 20, "Timeout connecting in seconds")

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

	fmt.Printf("Config: %+v\n", params)

	databaseClient := postgresql.DatabaseClient{
		Params: params.TimescaleDbParams,
	}

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
		dbClient:        databaseClient,
		clientset:       clientset,
		podQueue:        podQueue,
		nodeLister:      initInformers(clientset, podQueue, quit),
	}
}

func initInformers(clientset *kubernetes.Clientset, podQueue chan *v1.Pod, quit chan struct{}) listersv1.NodeLister {
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
			if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == os.Getenv("SCHEDULER_NAME") {
				podQueue <- pod
			}
		},
	})

	factory.Start(quit)
	return nodeInformer.Lister()
}

func main() {
	fmt.Println("Scheduler started!")

	rand.Seed(time.Now().Unix())

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

	filteredNodes := s.runPredicates(nodesToInspect, pod)
	if len(filteredNodes) == 0 {
		return "", errors.New("failed to find node that fits pod")
	}

	ipSlice := pkg.GetInternalIpsSlice(filteredNodes)

	priorityMap, _ := s.dbClient.GetMetrics(s.schedulerParams.MetricParams)

	var filteredPriorities = make(map[string]int)
	for k, v := range priorityMap {
		if pkg.Contains(ipSlice, k) {
			filteredPriorities[k] = v
		}
	}

	log.Println("calculated priorities after filter nodes where pod fit: ", filteredPriorities)

	bestNodeIp := s.findBestNode(filteredPriorities)
	bestNodeName := s.GetBestNodeName(filteredNodes, bestNodeIp)
	log.Println("bestNode", bestNodeName, " bestNodeIp:", bestNodeIp)
	return bestNodeName, nil
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
			Component: os.Getenv("SCHEDULER_NAME"),
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

func (s *Scheduler) runPredicates(nodes []*v1.Node, pod *v1.Pod) []*v1.Node {
	filteredNodes := make([]*v1.Node, 0)

	for _, node := range nodes {
		if s.fitResourcesPredicate(node, pod) {
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

func (s *Scheduler) fitResourcesPredicate(node *v1.Node, pod *v1.Pod) bool {

	var podCpu resource.Quantity
	var podMemory resource.Quantity

	for _, container := range pod.Spec.Containers {
		podCpu.Add(*container.Resources.Requests.Cpu())
		podMemory.Add(*container.Resources.Requests.Memory())
	}

	var nodeCpu resource.Quantity
	var nodeMem resource.Quantity

	pods, _ := s.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node.Name,
	})

	for _, npod := range pods.Items {
		for _, ncontainer := range npod.Spec.Containers {
			nodeCpu.Add(*ncontainer.Resources.Requests.Cpu())
			nodeMem.Add(*ncontainer.Resources.Requests.Memory())
		}
	}

	freeCpu := node.Status.Allocatable.Cpu()
	freeCpu.Sub(nodeCpu)

	freeMem := node.Status.Allocatable.Memory()
	freeMem.Sub(nodeMem)

	if freeCpu.Cmp(podCpu) == -1 || freeMem.Cmp(podMemory) == -1 {
		return false
	}

	return true
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

func (s *Scheduler) GetBestNodeName(nodes []*v1.Node, internalIp string) string {

	for _, node := range nodes {
		for _, address := range node.Status.Addresses {
			if string(address.Type) == "InternalIP" {
				if address.Address == internalIp {
					return node.Name
				}
			}
		}
	}
	return ""
}
