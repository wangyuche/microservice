# microservice

### POST JSON Parameter
| Parameter | Type | Describe |
| ------ | ------ | ------ |
| Req | string | 可以隨意輸入當回應值 |
| Http_fail_rate | int | 0~100 |
| Http_Status_code | int | 可用來當http status code |
| Http_Delay_rate | int | 0~100 |
| Http_Delay | int | ms |
| Next | string | 要去的下一個網址可不填寫 |

curl -X POST http://35.185.164.107:8080/v1/public/singlehttp -H 'Content-Type: application/json' -d '{"Req":"abc","Http_fail_rate":0,"Http_Status_code":200,"Http_Delay_rate":0,"Http_Delay":0,"CallMysql":1,"CallMongo":1}'

telepresence connect
telepresence quit
telepresence uninstall --everything
telepresence intercept -n aaa a1 --port 3000:3000