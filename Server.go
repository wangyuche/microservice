package main

import (
	c_rand "crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

var (
	commitid  = ""
	buildtime = ""
	version   = ""
)

func main() {
	app := fiber.New()
	_apiversion := app.Group("/v1")
	_public := _apiversion.Group("/public")
	_public.Post("/singlehttp", singlehttp)
	_public.Get("/appinfor", appinfor)
	err := app.Listen(":" + os.Getenv("port"))
	if err != nil {
		panic(err.Error())
	}
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
}
type NextCS struct {
	Req              string `json:"Req"`
	Http_fail_rate   int    `json:"Http_fail_rate"`
	Http_Status_code int    `json:"Http_Status_code"`
	Http_Delay_rate  int    `json:"Http_Delay_rate"`
	Http_Delay       int    `json:"Http_Delay"`
}

type SinglehttpSC struct {
	Res string `json:"Res"`
}

func singlehttp(c *fiber.Ctx) error {
	cs := new(SinglehttpCS)
	data := new(SinglehttpSC)
	if err := c.BodyParser(cs); err != nil {
		panic(err.Error())
	}
	result, err := c_rand.Int(c_rand.Reader, big.NewInt(int64(100)))
	if err != nil {
		panic(err.Error())
	}
	var _r int = int(result.Uint64())
	if _r <= cs.Http_fail_rate-1 {
		return c.Status(cs.Http_Status_code).SendString("")
	}

	result, err = c_rand.Int(c_rand.Reader, big.NewInt(int64(100)))
	if err != nil {
		panic(err.Error())
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
		bytes, err := json.Marshal(next)
		if err != nil {
			panic(err.Error())
		}
		req.SetBody(bytes)
		req.Header.SetContentType("application/json")
		req.Header.SetMethod("POST")
		resp := &fasthttp.Response{}
		client := &fasthttp.Client{}
		if err := client.Do(req, resp); err != nil {
			panic(err.Error())
		}
	}

	data.Res = cs.Req
	return c.JSON(data)
}
