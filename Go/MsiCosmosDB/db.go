package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/icrowley/fake"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type DB struct {
	connURI   string
	Container string
}

func NewDB(connURI, container string) *DB {
	return &DB{
		connURI:   connURI,
		Container: container,
	}
}

func (db *DB) getConn() *mgo.Session {
	dialInfo, err := mgo.ParseURL(db.connURI)
	if err != nil {
		log.Fatal(err)
	}

	// //Below part is similar to above.
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), &tls.Config{})
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatal(err)
	}

	session.SetSafe(&mgo.Safe{})

	return session
}

func (db *DB) InsertUsers(users []User) error {
	session := db.getConn()
	defer session.Close()

	c := session.DB(db.Container).C(db.Container)

	log.Println("Adding Recored to Databases")

	errOccured := false

	for idx := range users {
		//Container started the work
		err := c.Insert(&users[idx])
		if err != nil {
			errOccured = true
			log.Printf("Error while inserting user: %v", err)
		}
	}

	if errOccured {
		return fmt.Errorf("1 or more errors occured while inserting users into db")
	}

	return nil
}

func (db *DB) GetUsers() ([]User, error) {
	session := db.getConn()
	defer session.Close()

	c := session.DB(db.Container).C(db.Container)

	log.Println("Getting Users from Databases")

	var users []User
	err := c.Find(bson.M{}).All(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (db *DB) PopulateWithUsers(numUsers int) error {
	users := generateFakeUsers(numUsers)

	return db.InsertUsers(users)
}

func generateFakeUsers(num int) []User {
	users := make([]User, 0, num)
	for i := 0; i < num; i++ {
		users = append(users, User{
			Name:  fake.FirstName(),
			Email: fake.EmailAddress(),
		})
	}

	return users
}
