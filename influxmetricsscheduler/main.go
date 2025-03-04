package main

import (
	"flag"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"main/pkg/commons"
	"main/pkg/influxdb"
	"math"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
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
	schedulerParams commons.SchedulerParams
	dbClient        influxdb.DatabaseClient

	clientset  *kubernetes.Clientset
	podQueue   chan *v1.Pod
	nodeLister listersv1.NodeLister

	predicates []predicateFunc
}

func NewScheduler(podQueue chan *v1.Pod, quit chan struct{}) Scheduler {

	log.SetFormatter(&log.JSONFormatter{})
	var params = commons.SchedulerParams{}

	//metrics params
	flag.StringVar(&params.MetricParams.MetricName, "metric-name", commons.LookupEnvOrString("METRIC_NAME", ""), "Metric name in Prometheus to scheduled")
	flag.StringVar(&params.MetricParams.StartDate, "metric-start-date", commons.LookupEnvOrString("METRIC_START_DATE", "-1d"), "Start date to get metrics")
	flag.StringVar(&params.MetricParams.EndDate, "metric-end-date", commons.LookupEnvOrString("METRIC_END_DATE", "now()"), "End date to get metrics")
	flag.StringVar(&params.MetricParams.Operation, "metric-operation", commons.LookupEnvOrString("METRIC_OPERATION", "difference"), "Operation to get  metrics, example: max,min,avg,...")
	flag.StringVar(&params.MetricParams.PriorityOrder, "metric-priority-order", commons.LookupEnvOrString("METRIC_PRIORITY_ORDER", "desc"), "how to priority results, example. order asc o desc")
	flag.StringVar(&params.MetricParams.FilterClause, "metric-filter-clause", commons.LookupEnvOrString("METRIC_FILTER_CLAUSE", ""), "Extra filter clause")
	flag.BoolVar(&params.MetricParams.IsSecondLevel, "metric-is-second-level", commons.LookupEnvOrBool("METRIC_IS_SECOND_LEVEL", false), "Is second level")
	flag.StringVar(&params.MetricParams.SecondLevelGroup, "metric-second-level-group", commons.LookupEnvOrString("METRIC_SECOND_LEVEL_GROUP", ""), "Second level group")
	flag.StringVar(&params.MetricParams.SecondLevelOperation, "metric-second-level-operation", commons.LookupEnvOrString("METRIC_SECOND_LEVEL_OPERATION", ""), "Second level select")

	//TimescaleDbParams
	flag.StringVar(&params.Influxdb.Host, "influxdb-host", commons.LookupEnvOrString("INFLUXDB_HOST", "influxdb-influxdb2.monitoring"), "host to connect to influxdb")
	flag.StringVar(&params.Influxdb.Port, "influxdb-port", commons.LookupEnvOrString("INFLUXDB_PORT", "80"), "port to connect to influxdb")
	flag.StringVar(&params.Influxdb.Token, "influxdb-token", commons.LookupEnvOrString("INFLUXDB_TOKEN", "klsjdaioqwehrqoikdnmxcq"), "token to connect to influxdb")
	flag.StringVar(&params.Influxdb.Organization, "influxdb-organization", commons.LookupEnvOrString("INFLUXDB_ORGANIZATION", "uclm"), "organization where connect to influxdb")
	flag.StringVar(&params.Influxdb.Bucket, "influxdb-bucket", commons.LookupEnvOrString("INFLUXDB_BUCKET", "doctorado"), "bucket to connect to influxdb")

	flag.StringVar(&params.SchedulerName, "scheduler-name", commons.LookupEnvOrString("SCHEDULER_NAME", "random"), "scheduler name.")
	flag.StringVar(&params.LogLevel, "log-level", commons.LookupEnvOrString("LOG_LEVEL", "info"), "scheduler log level.")
	flag.StringVar(&params.FilteredNodes, "filtered-nodes", commons.LookupEnvOrString("FILTERED_NODES", ""), "Nodes to filer.")
	flag.IntVar(&params.Timeout, "timeout", commons.LookupEnvOrInt("TIMEOUT", 20), "Timeout connecting in seconds")

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

	fmt.Printf("Config: %v\n", params)

	databaseClient := influxdb.DatabaseClient{
		Params: params.Influxdb,
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
		nodeLister:      initInformers(clientset, podQueue, quit, params.SchedulerName),
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
	fmt.Println("Scheduler started!")

	rand.New(rand.NewSource(time.Now().Unix()))

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

	ipSlice := commons.GetInternalIpsSlice(filteredNodes)

	priorityMap, _ := s.dbClient.GetMetrics(s.schedulerParams.MetricParams)

	var filteredPriorities = make(map[string]int64)
	for k, v := range priorityMap {
		if commons.Contains(ipSlice, k) {
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

	log.Info("freeCpu: ", freeCpu, " freeMem: ", freeMem)
	log.Info("podCpu: ", podCpu, " podMemory: ", podMemory)
	if freeCpu.Cmp(podCpu) == -1 || freeMem.Cmp(podMemory) == -1 {
		return false
	}

	return true
}

func (s *Scheduler) findBestNode(priorities map[string]int64) string {
	var objectiveP int64
	var bestNode string

	log.Info("priorities: ", priorities)

	if s.schedulerParams.MetricParams.PriorityOrder == "asc" {
		objectiveP = 0
		for node, p := range priorities {
			if p > objectiveP {
				objectiveP = p
				bestNode = node
			}
		}
	} else {
		int64Max := int64(math.MaxInt64)
		objectiveP = int64Max
		for node, p := range priorities {
			if p < objectiveP {
				objectiveP = p
				bestNode = node
			}
		}
	}

	log.Info("bestNode: ", bestNode, " priority: ", objectiveP)

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
