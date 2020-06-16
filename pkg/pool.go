package pkg

import (
	"live/pkg/mqtt/pack"
	"sync"
)

/**
空的pubpack包资源池
*/
var pubPool *sync.Pool

func init() {
	pubPool = &sync.Pool{New: func() interface{} {
		return pack.NewEmptyPubPack()
	}}
}

func getEmptyPub() *pack.PubPack {
	return pubPool.Get().(*pack.PubPack)
}

func putEmptyPub(pubPack *pack.PubPack) {

	pubPack.SetEmpty()

	pubPool.Put(pubPack)

}
