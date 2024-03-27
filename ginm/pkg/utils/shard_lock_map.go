package utils

import (
	"sync"
)

var ShardCount = 32

type ShardMap struct {
	shards []*singleShard
	hash   IHash
}

type singleShard struct {
	items map[string]interface{}
	sync.RWMutex
}

func NewShardMap() *ShardMap {
	slm := &ShardMap{
		shards: make([]*singleShard, ShardCount),
		hash:   DefaultHash(),
	}
	for i := range slm.shards {
		slm.shards[i] = &singleShard{
			items:   make(map[string]interface{}),
			RWMutex: sync.RWMutex{},
		}
	}
	return slm
}

func (slm *ShardMap) GetShard(key string) *singleShard {
	return slm.shards[slm.hash.Sum(key)%uint32(ShardCount)]
}

func (slm *ShardMap) Set(key string, value interface{}) {
	shard := slm.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

func (slm *ShardMap) Count() int {
	count := 0
	for i := 0; i < ShardCount; i++ {
		slm.shards[i].Lock()
		count += len(slm.shards[i].items)
		slm.shards[i].Unlock()
	}
	return count
}

func (slm *ShardMap) Has(key string) bool {
	shard := slm.GetShard(key)
	shard.RLock()
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

func (slm *ShardMap) Get(key string) (interface{}, bool) {
	shard := slm.GetShard(key)
	shard.RLock()
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

type tuple struct {
	key   string
	value interface{}
}

func (slm *ShardMap) Iterators() <-chan tuple {
	chanList := snapshot(slm)
	total := 0
	for _, c := range chanList {
		total += cap(c)
	}
	ch := make(chan tuple, total)
	go fanIn(chanList, ch)
	return ch
}

func fanIn(chanList []chan tuple, out chan tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chanList))
	for _, c := range chanList {
		go func(ch chan tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(c)
	}
	wg.Wait()
	close(out)
}

func snapshot(slm *ShardMap) []chan tuple {
	chanList := make([]chan tuple, ShardCount)
	wg := sync.WaitGroup{}
	wg.Add(ShardCount)
	for i := 0; i < ShardCount; i++ {
		go func(index int, shard *singleShard) {
			chanList[index] = make(chan tuple, len(shard.items))
			wg.Done()
			shard.RWMutex.RLock()
			for key, value := range shard.items {
				chanList[index] <- tuple{
					key:   key,
					value: value,
				}
			}
			shard.RUnlock()
			close(chanList[index])
		}(i, slm.shards[i])
	}
	wg.Wait()
	return chanList
}

func (slm *ShardMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	for item := range slm.Iterators() {
		tmp[item.key] = item.value
	}

	return tmp
}

func (slm *ShardMap) Keys() []string {
	count := slm.Count()
	ch := make(chan string, count)
	go func() {
		wg := sync.WaitGroup{}
		wg.Add(ShardCount)
		for _, shard := range slm.shards {
			go func(shard *singleShard) {
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}
