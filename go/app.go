package main

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"redis"
	"bufio"
	"os"
	"io/ioutil"
	"runtime"
)

var loadtimeout uint64 = 0
var savetimeout uint64 = 0

func load(r redis.AsyncClient, k string, w http.ResponseWriter) (obj interface{}) {
	f, rerr := r.Get(k)
	if rerr != nil {
		panic(rerr)
	}
	val, rerr, timeout := f.TryGet(50000000000)
	if rerr != nil {
		panic(rerr)
	}
	if timeout {
		loadtimeout++
		log.Println("load timeout! count: ", loadtimeout)
		fmt.Fprintf(w, "Save failed for %s", key);
		return
	}
	zr, err := zlib.NewReader(bytes.NewReader(val))
	if err != nil {
		log.Fatal("Failed to create zlib reader with error: ", err)
	}
	defer zr.Close()
	jd := json.NewDecoder(zr)
	err = jd.Decode(&obj)
	if err != nil {
		log.Fatal("Failed to decode json with error: ", err)
	}
	return
}

func compute() {
	var k float64
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			k = math.Sin(float64(i)) * math.Sin(float64(j))
		}
	}
	_ = k
	//log.Println("Math Done")
}

func save(r redis.AsyncClient, key string, obj interface{}, w http.ResponseWriter) {
	var b bytes.Buffer

	z := zlib.NewWriter(&b)
	defer z.Close()

	je := json.NewEncoder(z)

	err := je.Encode(obj)
	if err != nil {
		log.Fatal("Failed to json Encode with error: ", err)
	}
	z.Flush()

	f, rerr := r.Set(key, b.Bytes())
	if rerr != nil {
		panic(rerr)
	}
	_, rerr, timeout := f.TryGet(50000000000)
	if rerr != nil {
		panic(rerr)
	}
	if timeout {
		savetimeout++
		log.Println("save timeout! count: ", savetimeout)
		fmt.Fprintf(w, "Save failed for %s", key);
	}
}

type User struct {
	Id      uint64
	Name    string
	Xp      uint64
	Level   uint64
	Items   []Item
	Friends []uint64
}
type Item struct {
	Id   uint64
	Type string
	Data string
}

func NewUser() *User {
	it := make([]Item, 20)
	for i, _ := range it {
		it[i].Id = 1000 + uint64(i)
		it[i].Type = "sometype"
		it[i].Data = "some data blah blah blah"
	}
	friends := make([]uint64, 50)
	for i, _ := range friends {
		friends[i] = uint64(i) + 10000000
	}
	user := &User{
		Id:      1292983,
		Name:    "Vinay",
		Xp:      100,
		Level:   200,
		Items:   it,
		Friends: friends,
	}
	return user
}

var obj interface{}
var key string

func responseHandler(w http.ResponseWriter, r *http.Request, client redis.AsyncClient) {
	obj = load(client, key, w)
	//log.Print(obj)
	compute()
	//obj = NewUser()
	save(client, key, obj, w)
	fmt.Fprintf(w, "OK! %s", key)
}

func main() {
	runtime.GOMAXPROCS(16)
	key = "user_data_2"
	spec := redis.DefaultSpec().Db(0).Host("10.174.178.235")
	client, err := redis.NewAsynchClientWithSpec(spec)
	if err != nil {
		panic(err)
	}
	primeKey(key, client)
	defer client.Quit()
	handler := func(w http.ResponseWriter, r *http.Request) {
		responseHandler(w, r, client)
	}
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}

func primeKey(key string, r redis.AsyncClient){
		path := "document.json";
	    file, err := os.Open(path);
	    if err != nil {
	    	panic(err)
	    }
        reader := bufio.NewReader(file)
		document, _ := ioutil.ReadAll(reader)
		var b bytes.Buffer
		z := zlib.NewWriter(&b)
		z.Write(document)
		z.Close();
		f, rerr := r.Set(key, b.Bytes())
		if rerr != nil {
			panic(rerr)
		}
		_, rerr, timeout := f.TryGet(50000000000)
		if rerr != nil {
			panic(rerr)
		}
		if timeout {
			savetimeout++
			log.Println("save timeout! count: ", savetimeout)
		}
}