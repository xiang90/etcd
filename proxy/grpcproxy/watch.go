package grpcproxy

import (
	"io"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/clientv3"
	pb "github.com/coreos/etcd/etcdserver/etcdserverpb"
)

type watchProxy struct {
	c clientv3.Client

	watchChs map[watchRange]coalescedWatcher
}

type watchRange struct {
	key string
	end string
}

func (wp *watchProxy) Watch(stream pb.Watch_WatchServer) error {

}

type watchStream struct {
	serverStream pb.Watch_WatchServer

	wc clientv3.Watcher
}

func (ws watchStream) recvLoop() error {
	for {
		r, err := ws.serverStream()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if create := r.GetCreateRequest(); create != nil {
			wchan, err := ws.wc.Watch(context.TODO(), string(create.Key))
		}
		if cancel := r.GetCancelRequest(); cancel != nil {
		}
	}
}

type coalescedWatcher struct {
	wc      clientv3.WatchChan
	streams map[pb.Watch_WatchServer]struct{}
}

func (cw *coalescedWatcher) run() {
	for wr := range <-cw.wc {
		for s := range cw.streams {
			s.Send()
		}
	}
}
