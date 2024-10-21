nodes=$(kubectl get nodes -o name | cut -d '/' -f 2-)

for node in ${nodes[@]}
do
    echo "==== Shut down $node ===="
    ssh $node sudo shutdown -h 1
done
