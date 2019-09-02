package mongodb

import (
	"config"
	"github.com/globalsign/mgo"
	"log"
	"time"
)

var session *mgo.Session

type SessionStore struct {
	session *mgo.Session
}

// 初始化
func init() {
	// 连接数据库
	var err error
	session, err = mgo.DialWithTimeout(config.MongoConf["connect"], 5*time.Second)
	// 异常处理
	if err != nil {
		log.Panic(err)
	}
	//
	session.SetMode(mgo.Monotonic, true)
}

// 获取连接
func GetS() *SessionStore {
	return &SessionStore{
		session: session.Copy(),
	}
}

// 集合实例
func (s *SessionStore) GetC(cName string) *mgo.Collection {
	return s.session.DB(config.MongoConf["database"]).C(cName)
}

// 关闭连接
func (s *SessionStore) Close() {
	s.session.Close()
}
