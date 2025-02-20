package cmd

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtimer"
	"log"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/os/gcmd"
	"plat_order/internal/service"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			lao := service.ListenAndOrder()

			err = lao.SetSymbol(ctx)
			if nil != err {
				log.Println("启动错误，币种信息：", err)
			}

			// 300秒/次，币种信息
			handle := func(ctx context.Context) {
				err = lao.SetSymbol(ctx)
				if nil != err {
					log.Println("任务错误，币种信息：", err)
				}
			}
			gtimer.AddSingleton(ctx, time.Minute*5, handle)

			err = lao.PullAndSetTraderUserPositionSide(ctx)
			if nil != err {
				log.Println("启动错误，同步交易员和用户持仓方向：", err)
			}

			//// 1分钟/次，同步持仓信息和持仓方向
			//handle2 := func(ctx context.Context) {
			//	err = lao.PullAndSetTraderUserPositionSide(ctx)
			//	if nil != err {
			//		log.Println("任务错误，同步交易员和用户持仓方向：", err)
			//	}
			//}
			//gtimer.AddSingleton(ctx, time.Minute*1, handle2)

			lao.PullAndSetBaseMoneyNewGuiTuAndUser(ctx)
			// 1分钟/次，同步持仓信息和持仓方向
			handle3 := func(ctx context.Context) {
				lao.PullAndSetBaseMoneyNewGuiTuAndUser(ctx)
			}
			gtimer.AddSingleton(ctx, time.Minute*1, handle3)

			// 30秒/次，更新用户信息 todo
			handle4 := func(ctx context.Context) {
				err = lao.SetUser(ctx)
				if nil != err {
					log.Println("任务错误，设置用户：", err)
				}
			}
			gtimer.AddSingleton(ctx, time.Second*30, handle4)

			//// 5分钟/次，更新用户信息 todo
			//handle5 := func(ctx context.Context) {
			//	lao.HandleBothPositions(ctx)
			//}
			//gtimer.AddSingleton(ctx, time.Minute*5, handle5)

			// 启动
			go lao.Run(ctx)

			// 开启http管理服务
			s := g.Server()
			s.Group("/api", func(group *ghttp.RouterGroup) {
				// 探测ip
				group.POST("/set_position_side", func(r *ghttp.Request) {
					var (
						setCode   uint64
						setString string
					)

					setCode, setString = lao.SetPositionSide(r.PostFormValue("api_key"), r.PostFormValue("api_secret"))
					r.Response.WriteJson(g.Map{
						"code": setCode,
						"msg":  setString,
					})

					return
				})

				// 查询num
				group.GET("/nums", func(r *ghttp.Request) {
					res := lao.GetSystemUserNum(ctx)

					responseData := make([]*g.MapStrAny, 0)
					for k, v := range res {
						responseData = append(responseData, &g.MapStrAny{k: v})
					}

					r.Response.WriteJson(responseData)
					return
				})

				// 更新num
				group.POST("/update/num", func(r *ghttp.Request) {
					var (
						parseErr error
						setErr   error
						num      float64
					)
					num, parseErr = strconv.ParseFloat(r.PostFormValue("num"), 64)
					if nil != parseErr || 0 >= num {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					setErr = lao.SetSystemUserNum(ctx, r.PostFormValue("apiKey"), num)
					if nil != setErr {
						r.Response.WriteJson(g.Map{
							"code": -2,
						})

						return
					}

					r.Response.WriteJson(g.Map{
						"code": 1,
					})

					return
				})

				// 加人
				group.POST("/create/user", func(r *ghttp.Request) {
					var (
						parseErr error
						setErr   error
						needInit uint64
						num      float64
					)
					needInit, parseErr = strconv.ParseUint(r.PostFormValue("need_init"), 10, 64)
					if nil != parseErr {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					num, parseErr = strconv.ParseFloat(r.PostFormValue("num"), 64)
					if nil != parseErr || 0 >= num {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					setErr = lao.CreateUser(
						ctx,
						r.PostFormValue("address"),
						r.PostFormValue("api_key"),
						r.PostFormValue("api_secret"),
						"binance",
						needInit,
						num,
					)
					if nil != setErr {
						r.Response.WriteJson(g.Map{
							"code": -2,
						})

						return
					}

					r.Response.WriteJson(g.Map{
						"code": 1,
					})

					return
				})

				// 更新api status
				group.POST("/update/api_status", func(r *ghttp.Request) {
					var (
						parseErr error
						setCode  uint64
						status   uint64
						reInit   uint64
					)
					status, parseErr = strconv.ParseUint(r.PostFormValue("status"), 10, 64)
					if nil != parseErr || 0 >= status {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					reInit, parseErr = strconv.ParseUint(r.PostFormValue("init"), 10, 64)
					if nil != parseErr || 0 >= reInit {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					setCode = lao.SetApiStatus(ctx, r.PostFormValue("apiKey"), status, reInit)
					r.Response.WriteJson(g.Map{
						"code": setCode,
					})

					return
				})

				// 更新开新单
				group.POST("/update/useNewSystem", func(r *ghttp.Request) {
					var (
						parseErr error
						setErr   error
						status   uint64
					)
					status, parseErr = strconv.ParseUint(r.PostFormValue("status"), 10, 64)
					if nil != parseErr || 0 > status {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					setErr = lao.SetUseNewSystem(ctx, r.PostFormValue("apiKey"), status)
					if nil != setErr {
						r.Response.WriteJson(g.Map{
							"code": -2,
						})

						return
					}

					r.Response.WriteJson(g.Map{
						"code": 1,
					})

					return
				})

				// 查询用户系统仓位
				group.GET("/user/positions", func(r *ghttp.Request) {
					res := lao.GetSystemUserPositions(ctx, r.Get("apiKey").String())

					responseData := make([]*g.MapStrAny, 0)
					for k, v := range res {
						responseData = append(responseData, &g.MapStrAny{k: v})
					}

					r.Response.WriteJson(responseData)
					return
				})

				// 查询用户binance仓位
				group.GET("/user/binance/positions", func(r *ghttp.Request) {
					res := lao.GetBinanceUserPositions(ctx, r.Get("apiKey").String())

					responseData := make([]*g.MapStrAny, 0)
					for k, v := range res {
						responseData = append(responseData, &g.MapStrAny{k: v})
					}

					r.Response.WriteJson(responseData)
					return
				})

				// 用户全平仓位
				group.POST("/user/close/positions", func(r *ghttp.Request) {
					r.Response.WriteJson(g.Map{
						"code": lao.CloseBinanceUserPositions(ctx),
					})

					return
				})

				// 用户设置仓位
				group.POST("/user/update/position", func(r *ghttp.Request) {
					var (
						parseErr     error
						num          float64
						system       uint64
						allCloseGate uint64
					)
					num, parseErr = strconv.ParseFloat(r.PostFormValue("num"), 64)
					if nil != parseErr || 0 >= num {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					system, parseErr = strconv.ParseUint(r.PostFormValue("system"), 10, 64)
					if nil != parseErr || 0 > system {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					allCloseGate, parseErr = strconv.ParseUint(r.PostFormValue("allCloseGate"), 10, 64)
					if nil != parseErr || 0 > allCloseGate {
						r.Response.WriteJson(g.Map{
							"code": -1,
						})

						return
					}

					r.Response.WriteJson(g.Map{
						"code": lao.SetSystemUserPosition(
							ctx,
							system,
							allCloseGate,
							r.PostFormValue("apiKey"),
							r.PostFormValue("symbol"),
							r.PostFormValue("side"),
							r.PostFormValue("positionSide"),
							num,
						),
					})

					return
				})
			})

			s.SetPort(80)
			s.Run()

			return nil
		},
	}
)
