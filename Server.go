package main

import (
	"context"
	c_rand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	redis "github.com/go-redis/redis/v8"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/qiniu/qmgo"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

var (
	commitid  = ""
	buildtime = ""
	version   = ""
)
var isshutdown bool = false
var ctx = context.Background()
var rdb *redis.ClusterClient
var qmgoDB *qmgo.Database
var ctxbg = context.Background()
var connectTimeoutMS = int64(60 * time.Second / time.Millisecond)
var pool uint64 = 1

func main() {
	SetReadConnectionInfo("test", os.Getenv("mysql"), 1, 1)
	rdb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{"redis.arieswang:6379"},
		Password:     "", // no password set
		PoolSize:     1,
		MinIdleConns: 1,
		MaxRetries:   10,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
	})
	cfg := &qmgo.Config{
		Uri:              os.Getenv("mongo"),
		ConnectTimeoutMS: &connectTimeoutMS,
		Database:         "test",
		MaxPoolSize:      &pool,
		MinPoolSize:      &pool,
		Auth: &qmgo.Credential{
			Username: "root",
			Password: "yile.net",
		},
	}
	client, err := qmgo.Open(ctxbg, cfg)
	if err != nil {
		panic(err.Error())
	}
	qmgoDB = client.Database
	app := fiber.New()
	_apiversion := app.Group("/v1")
	_public := _apiversion.Group("/public")
	_private := _apiversion.Group("/private")
	_public.Post("/singlehttp", singlehttp)
	_public.Get("/appinfor", appinfor)
	_private.Get("/hc", hc)
	go func() {
		err := app.Listen(":" + os.Getenv("port"))
		if err != nil {
			panic(err.Error())
		}
	}()
	go func() {
		app1 := fiber.New()
		app1.Get("/", func(c *fiber.Ctx) error {
			return c.SendString("Hello, World 👋!")
		})
		app1.Listen(":3000")
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	_ = <-quit
	isshutdown = true
	time.Sleep(5 * time.Second)
	app.Shutdown()
}

func appinfor(c *fiber.Ctx) error {
	msg := fmt.Sprintf("CommitId:%s \n", commitid)
	msg = fmt.Sprintf(msg+"BuildTime:%s \n", buildtime)
	msg = fmt.Sprintf(msg+"Version:%s \n", version)
	return c.SendString(msg)
}

type SinglehttpCS struct {
	Req              string `json:"Req"`
	Http_fail_rate   int    `json:"Http_fail_rate"`
	Http_Status_code int    `json:"Http_Status_code"`
	Http_Delay_rate  int    `json:"Http_Delay_rate"`
	Http_Delay       int    `json:"Http_Delay"`
	Next             string `json:"Next,omitempty"`
	CallRedis        int    `json:"CallRedis,omitempty"`
	NextCallRedis    int    `json:"NextCallRedis,omitempty"`
	CallMysql        int    `json:"CallMysql,omitempty"`
	NextCallMysql    int    `json:"NextCallMysql,omitempty"`
	CallMongo        int    `json:"CallMongo,omitempty"`
	NextCallMongo    int    `json:"NextCallMongo,omitempty"`
}
type NextCS struct {
	Req              string `json:"Req"`
	Http_fail_rate   int    `json:"Http_fail_rate"`
	Http_Status_code int    `json:"Http_Status_code"`
	Http_Delay_rate  int    `json:"Http_Delay_rate"`
	Http_Delay       int    `json:"Http_Delay"`
	CallRedis        int    `json:"CallRedis,omitempty"`
	CallMysql        int    `json:"CallMysql,omitempty"`
	CallMongo        int    `json:"CallMongo,omitempty"`
}

type SinglehttpSC struct {
	Res string `json:"Res"`
}
type test struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
}

func singlehttp(c *fiber.Ctx) error {
	cs := new(SinglehttpCS)
	data := new(SinglehttpSC)
	if err := c.BodyParser(cs); err != nil {
		fmt.Println("err", err.Error())
		return c.Status(500).SendString("")
	}
	result, err := c_rand.Int(c_rand.Reader, big.NewInt(int64(100)))
	if err != nil {
		fmt.Println("err", err.Error())
		return c.Status(500).SendString("")
	}
	var _r int = int(result.Uint64())
	if _r <= cs.Http_fail_rate-1 {
		return c.Status(cs.Http_Status_code).SendString("")
	}

	result, err = c_rand.Int(c_rand.Reader, big.NewInt(int64(100)))
	if err != nil {
		fmt.Println("err", err.Error())
		return c.Status(500).SendString("")
	}
	_r = int(result.Uint64())
	if _r <= cs.Http_Delay_rate-1 {
		time.Sleep(time.Duration(cs.Http_Delay) * time.Millisecond)
	}

	if cs.Next != "" {
		next := new(SinglehttpCS)
		req := &fasthttp.Request{}
		req.SetRequestURI(cs.Next)
		next.Http_Delay = cs.Http_Delay
		next.Http_Delay_rate = cs.Http_Delay_rate
		next.Http_Status_code = cs.Http_Status_code
		next.Http_fail_rate = cs.Http_fail_rate
		next.Req = cs.Req
		if cs.NextCallRedis == 1 {
			next.CallRedis = cs.NextCallRedis
		}
		if cs.NextCallMysql == 1 {
			next.CallMysql = cs.NextCallMysql
		}
		if cs.NextCallMongo == 1 {
			next.CallMongo = cs.NextCallMongo
		}
		bytes, err := json.Marshal(next)
		if err != nil {
			fmt.Println("err", err.Error())
			return c.Status(500).SendString("")
		}
		req.SetBody(bytes)
		req.Header.SetContentType("application/json")
		req.Header.SetMethod("POST")
		resp := &fasthttp.Response{}
		client := &fasthttp.Client{}
		if err := client.Do(req, resp); err != nil {
			fmt.Println("err", err.Error())
			return c.Status(500).SendString("")
		}
	}
	if cs.CallRedis == 1 {
		err := rdb.Set(ctx, "key", "value", 0).Err()
		if err != nil {
			fmt.Println("err", err.Error())
			return c.Status(500).SendString("")
		}

		_, err = rdb.Get(ctx, "key").Result()
		if err != nil {
			fmt.Println("err", err.Error())
			return c.Status(500).SendString("")
		}
	}
	if cs.CallMysql == 1 {
		conn, err := GetReadConnection("test").Begin()
		if err != nil {
			fmt.Println("err", err.Error())
			return c.Status(500).SendString("")
		}
		defer conn.Commit()
		sql := "INSERT INTO test(test) VALUES ('abc')"
		_, err = conn.Exec(sql)
		if err != nil {
			fmt.Println("err", err.Error())
			return c.Status(500).SendString("")
		}
	}
	if cs.CallMongo == 1 {
		_, err = qmgoDB.Collection("test").InsertOne(ctxbg, &test{})
		if err != nil {
			fmt.Println("err", err.Error())
			return c.Status(500).SendString("")
		}
		find := bson.M{}
		find["type"] = bson.M{"$in": []int{1}}
		if _, err = qmgoDB.Collection("test").Find(ctxbg, find).Count(); err != nil {
			fmt.Println("err", err.Error())
			return c.Status(500).SendString("")
		}
	}
	data.Res = cs.Req
	return c.JSON(data)
}

func hc(c *fiber.Ctx) error {
	if isshutdown {
		return c.SendStatus(404)
	} else {
		return c.SendStatus(200)
	}
}
