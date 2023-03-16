for i in $(ls); do
    if [ ! -f $i ]; then
        cd $i && 
        go build && 
        cp $i ../../kbin && 
        mv $i ${HOME}/.kbin && cd ..;
    fi
done
