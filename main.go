package main

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var mongoConn *mgo.Session

type MyEntity struct {
	Data []byte `json:"data" bson:"data"`
}

func createConnection() (*mgo.Session, error) {
	dialInfo := mgo.DialInfo{
		Addrs: []string{
			"abc-shard-00-00.gcp.mongodb.net:27017",
			"abc-shard-00-01.gcp.mongodb.net:27017",
			"abc-shard-00-02.gcp.mongodb.net:27017"},
		Username: "MongoUser",
		Password: "YourVerySecurePassword",
	}
	tlsConfig := &tls.Config{}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}
	return mgo.DialWithInfo(&dialInfo)
}

func main() {
	var err error
	mongoConn, err = createConnection()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/save", post)
	http.HandleFunc("/read", get)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func post(w http.ResponseWriter, req *http.Request) {
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	session := mongoConn.Copy()
	defer session.Close()

	entity := MyEntity{Data: payload}
	err = session.DB("test").C("data").Insert(entity)
	if err != nil {
		panic(err)
	}
}

func get(w http.ResponseWriter, req *http.Request) {
	session := mongoConn.Copy()
	defer session.Close()

	entity := MyEntity{}
	err := session.DB("test").C("data").Find(bson.M{}).One(&entity)
	if err != nil {
		panic(err)
	}

	w.Write(entity.Data)
	w.Write([]byte{10})
}
