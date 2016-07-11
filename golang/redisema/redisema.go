package redisema

import (
	"errors"
	"fmt"
	"time"

	"github.com/dailing/levlog"
	"github.com/garyburd/redigo/redis"
)

/*
 * Must keep keys in different client the same one
 * So no hash or other process on name,
 * just add a complex prefix to it.
 */
const __PREFIX__ = "__REDIS_SEMAPHORE_KEY__"

type RediSemaphore struct {
	redisDB *redis.Pool
	name    string
}

func NewResiSemaphore(name string) *RediSemaphore {
	return &RediSemaphore{
		redisDB: &redis.Pool{
			MaxIdle:     100,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", "localhost:6379")
				if err != nil {
					return nil, err
				}
				return c, err
			},
		},
		name: name,
	}
}

func (r *RediSemaphore) Init(initNum int) error {
	Inited, err := r.CheckInited()
	levlog.E(err)
	if err != nil {
		return err
	}
	if Inited {
		return errors.New(fmt.Sprint("Semaphore ", r.name, " already inited"))
	}
	// perform the init Now
	conn := r.redisDB.Get()
	defer conn.Close()
	err = conn.Send("RPUSH", r.getKeyName(), "1")
	levlog.E(err)
	err = conn.Send("LPOP", r.getKeyName())
	levlog.E(err)
	for i := 0; i < initNum; i++ {
		conn.Send("RPUSH", r.getKeyName(), "1")
	}
	err = conn.Flush()
	return err
}

func (r *RediSemaphore) Remove() error {
	conn := r.redisDB.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", r.getKeyName())
	return err
}

func (r *RediSemaphore) CheckInited() (bool, error) {
	conn := r.redisDB.Get()
	defer conn.Close()
	result, err := redis.Int(conn.Do("EXISTS", r.getKeyName()))
	if result != 0 {
		return true, err
	}
	return false, err
}

func (r *RediSemaphore) getKeyName() string {
	return __PREFIX__ + r.name
}

func (r *RediSemaphore) Wait() {
	conn := r.redisDB.Get()
	defer conn.Close()
	_, err := conn.Do("BLPOP", r.getKeyName(), 0)
	levlog.E(err)
}

func (r *RediSemaphore) Signal() {
	conn := r.redisDB.Get()
	defer conn.Close()
	_, err := conn.Do("RPUSH", r.getKeyName(), "1")
	levlog.E(err)
}
