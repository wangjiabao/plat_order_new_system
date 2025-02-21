package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"plat_order/internal/model/entity"
	"plat_order/internal/service"
	"strconv"
	"strings"
	"time"
)

type (
	sBinance struct{}
)

func init() {
	service.RegisterBinance(New())
}

func New() *sBinance {
	return &sBinance{}
}

const (
	apiBaseURL   = "https://fapi.binance.com"
	wsBaseURL    = "wss://fstream.binance.com/ws/"
	listenKeyURL = "/fapi/v1/listenKey"
)

// 获取币安服务器时间
func getBinanceServerTime() int64 {
	urlTmp := "https://api.binance.com/api/v3/time"
	resp, err := http.Get(urlTmp)
	if err != nil {
		log.Println("Error getting server time:", err)
		return 0
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	var serverTimeResponse struct {
		ServerTime int64 `json:"serverTime"`
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return 0
	}
	if err := json.Unmarshal(body, &serverTimeResponse); err != nil {
		log.Println("Error unmarshaling server time:", err)
		return 0
	}

	return serverTimeResponse.ServerTime
}

// 生成签名
func generateSignature(apiS string, params url.Values) string {
	// 将请求参数编码成 URL 格式的字符串
	queryString := params.Encode()

	// 生成签名
	mac := hmac.New(sha256.New, []byte(apiS))
	mac.Write([]byte(queryString)) // 用 API Secret 生成签名
	return hex.EncodeToString(mac.Sum(nil))
}

// GetBinancePositionSide 获取账户信息
func (s *sBinance) GetBinancePositionSide(apiK, apiS string) string {
	// 请求的API地址
	endpoint := "/fapi/v1/positionSide/dual"
	baseURL := "https://fapi.binance.com"

	// 获取当前时间戳（使用服务器时间避免时差问题）
	serverTime := getBinanceServerTime()
	if serverTime == 0 {
		return ""
	}
	timestamp := strconv.FormatInt(serverTime, 10)

	// 设置请求参数
	params := url.Values{}
	params.Set("timestamp", timestamp)
	params.Set("recvWindow", "5000") // 设置接收窗口

	// 生成签名
	signature := generateSignature(apiS, params)

	// 将签名添加到请求参数中
	params.Set("signature", signature)

	// 构建完整的请求URL
	requestURL := baseURL + endpoint + "?" + params.Encode()

	// 创建请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return ""
	}

	// 添加请求头
	req.Header.Add("X-MBX-APIKEY", apiK)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return ""
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return ""
	}

	// 解析响应
	var o *entity.PositionSide
	err = json.Unmarshal(body, &o)
	if err != nil {
		log.Println("Error unmarshalling response:", err)
		return ""
	}

	res := ""
	if o.DalSidePosition {
		res = "ALL"
	} else {
		res = "BOTH"
	}

	return res
}

// GetLatestPrice 获取价格
func (s *sBinance) GetLatestPrice(symbol string) string {
	baseURL := "https://api.binance.com/api/v3/ticker/price"
	query := url.Values{}
	query.Add("symbol", symbol)

	// 构建请求 URL
	requestURL := baseURL + "?" + query.Encode()
	resp, err := http.Get(requestURL)
	if err != nil {
		log.Println("获取价格错误：", err)
		return ""
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	// 读取响应数据
	var data *entity.LatestPrice
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Println("解析 JSON 错误：", err)
		return ""
	}

	if data == nil {
		log.Println("解析结果为空")
		return ""
	}

	return data.Price
}

// GetWalletInfo 获取钱包信息
func (s *sBinance) GetWalletInfo(apiK, apiS string) []*entity.WalletInfo {
	// 请求的API地址
	endpoint := "/sapi/v1/asset/wallet/balance"
	baseURL := "https://api.binance.com"

	res := make([]*entity.WalletInfo, 0)
	// 获取当前时间戳（使用服务器时间避免时差问题）
	serverTime := getBinanceServerTime()
	if serverTime == 0 {
		return res
	}
	timestamp := strconv.FormatInt(serverTime, 10)

	// 设置请求参数
	params := url.Values{}
	params.Set("timestamp", timestamp)
	params.Set("recvWindow", "5000") // 设置接收窗口

	// 生成签名
	signature := generateSignature(apiS, params)

	// 将签名添加到请求参数中
	params.Set("signature", signature)

	// 构建完整的请求URL
	requestURL := baseURL + endpoint + "?" + params.Encode()

	// 创建请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return res
	}

	// 添加请求头
	req.Header.Add("X-MBX-APIKEY", apiK)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return res
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return res
	}

	// 解析响应\
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Println("Error unmarshalling response:", err)
		return res
	}

	// 返回资产余额
	return res
}

// GetBinanceInfo 获取账户信息
func (s *sBinance) GetBinanceInfo(apiK, apiS string) string {
	// 请求的API地址
	endpoint := "/fapi/v2/account"
	baseURL := "https://fapi.binance.com"

	// 获取当前时间戳（使用服务器时间避免时差问题）
	serverTime := getBinanceServerTime()
	if serverTime == 0 {
		return ""
	}
	timestamp := strconv.FormatInt(serverTime, 10)

	// 设置请求参数
	params := url.Values{}
	params.Set("timestamp", timestamp)
	params.Set("recvWindow", "5000") // 设置接收窗口

	// 生成签名
	signature := generateSignature(apiS, params)

	// 将签名添加到请求参数中
	params.Set("signature", signature)

	// 构建完整的请求URL
	requestURL := baseURL + endpoint + "?" + params.Encode()

	// 创建请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return ""
	}

	// 添加请求头
	req.Header.Add("X-MBX-APIKEY", apiK)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return ""
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return ""
	}

	// 解析响应
	var o *entity.Asset
	err = json.Unmarshal(body, &o)
	if err != nil {
		log.Println("Error unmarshalling response:", err)
		return ""
	}

	// 返回资产余额
	return o.TotalMarginBalance
}

func (s *sBinance) RequestBinancePositionSide(positionSide string, apiKey string, secretKey string) (error, string, bool) {
	var (
		client       *http.Client
		req          *http.Request
		resp         *http.Response
		resOrderInfo *entity.BinanceOrderInfo
		data         string
		b            []byte
		err          error
		apiUrl       = "https://fapi.binance.com/fapi/v1/positionSide/dual"
	)

	//log.Println(symbol, side, orderType, positionSide, quantity, apiKey, secretKey)
	// 时间
	now := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	// 拼请求数据
	data = "dualSidePosition=" + positionSide + "&timestamp=" + now

	// 加密
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))
	// 构造请求

	req, err = http.NewRequest("POST", apiUrl, strings.NewReader(data+"&signature="+signature))
	if err != nil {
		return err, "", false
	}
	// 添加头信息
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-MBX-APIKEY", apiKey)

	// 请求执行
	client = &http.Client{Timeout: 3 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		return err, "", false
	}

	// 结果
	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Println(string(b), err)
		return err, string(b), false
	}

	err = json.Unmarshal(b, &resOrderInfo)
	if err != nil {
		//log.Println(string(b), err)
		return err, string(b), false
	}

	//log.Println(string(b))
	if 200 == resOrderInfo.Code || -4059 == resOrderInfo.Code {
		return nil, string(b), true
	}

	//log.Println(string(b), err)
	return nil, string(b), false
}

// GetBinanceFuturesPairs 获取 Binance U 本位合约交易对信息
func (s *sBinance) GetBinanceFuturesPairs() ([]*entity.BinanceSymbolInfo, error) {
	apiUrl := "https://fapi.binance.com/fapi/v1/exchangeInfo"

	// 发送 HTTP GET 请求
	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析 JSON 响应
	var exchangeInfo *entity.BinanceExchangeInfoResp
	err = json.Unmarshal(body, &exchangeInfo)
	if err != nil {
		return nil, err
	}

	return exchangeInfo.Symbols, nil
}

// RequestBinanceOrder 请求下单
func (s *sBinance) RequestBinanceOrder(symbol string, side string, orderType string, positionSide string, quantity string, apiKey string, secretKey string, reduceOnly bool) (*entity.BinanceOrder, *entity.BinanceOrderInfo, error) {
	var (
		client       *http.Client
		req          *http.Request
		resp         *http.Response
		res          *entity.BinanceOrder
		resOrderInfo *entity.BinanceOrderInfo
		data         string
		b            []byte
		err          error
		apiUrl       = "https://fapi.binance.com/fapi/v1/order"
	)

	//log.Println(symbol, side, orderType, positionSide, quantity, apiKey, secretKey)
	// 时间
	now := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	// 拼请求数据
	if reduceOnly {
		data = "symbol=" + symbol + "&side=" + side + "&type=" + orderType + "&positionSide=" + positionSide + "&newOrderRespType=" + "RESULT" + "&reduceOnly=true&quantity=" + quantity + "&timestamp=" + now
	} else {
		data = "symbol=" + symbol + "&side=" + side + "&type=" + orderType + "&positionSide=" + positionSide + "&newOrderRespType=" + "RESULT" + "&quantity=" + quantity + "&timestamp=" + now
	}

	// 加密
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	signature := hex.EncodeToString(h.Sum(nil))
	// 构造请求

	req, err = http.NewRequest("POST", apiUrl, strings.NewReader(data+"&signature="+signature))
	if err != nil {
		return nil, nil, err
	}
	// 添加头信息
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-MBX-APIKEY", apiKey)

	// 请求执行
	client = &http.Client{Timeout: 3 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	// 结果
	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(string(b), err)
		return nil, nil, err
	}

	var o *entity.BinanceOrder
	err = json.Unmarshal(b, &o)
	if err != nil {
		log.Println(string(b), err)
		return nil, nil, err
	}

	res = &entity.BinanceOrder{
		OrderId:       o.OrderId,
		ExecutedQty:   o.ExecutedQty,
		ClientOrderId: o.ClientOrderId,
		Symbol:        o.Symbol,
		AvgPrice:      o.AvgPrice,
		CumQuote:      o.CumQuote,
		Side:          o.Side,
		PositionSide:  o.PositionSide,
		ClosePosition: o.ClosePosition,
		Type:          o.Type,
	}

	if 0 >= res.OrderId {
		//log.Println(string(b))
		err = json.Unmarshal(b, &resOrderInfo)
		if err != nil {
			log.Println(string(b), err)
			return nil, nil, err
		}
	}

	return res, resOrderInfo, nil
}

// GetBinancePositionInfo 获取账户信息
func (s *sBinance) GetBinancePositionInfo(apiK, apiS string) []*entity.BinancePosition {
	// 请求的API地址
	endpoint := "/fapi/v2/account"
	baseURL := "https://fapi.binance.com"

	// 获取当前时间戳（使用服务器时间避免时差问题）
	serverTime := getBinanceServerTime()
	if serverTime == 0 {
		return nil
	}
	timestamp := strconv.FormatInt(serverTime, 10)

	// 设置请求参数
	params := url.Values{}
	params.Set("timestamp", timestamp)
	params.Set("recvWindow", "5000") // 设置接收窗口

	// 生成签名
	signature := generateSignature(apiS, params)

	// 将签名添加到请求参数中
	params.Set("signature", signature)

	// 构建完整的请求URL
	requestURL := baseURL + endpoint + "?" + params.Encode()

	// 创建请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return nil
	}

	// 添加请求头
	req.Header.Add("X-MBX-APIKEY", apiK)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return nil
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return nil
	}

	// 解析响应
	var o *entity.BinanceResponse
	err = json.Unmarshal(body, &o)
	if err != nil {
		log.Println("Error unmarshalling response:", err)
		return nil
	}

	// 返回资产余额
	return o.Positions
}

var (
	ListenKey = gvar.New("", true)
	Conn      *websocket.Conn // WebSocket connection
)

// ListenKeyResponse represents the response from Binance API when creating or renewing a ListenKey
type ListenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

// CreateListenKey creates a new ListenKey for user data stream
func (s *sBinance) CreateListenKey(apiKey string) error {
	req, err := http.NewRequest("POST", apiBaseURL+listenKeyURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return gerror.Newf("API error: %s", string(body))
	}

	var response *ListenKeyResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	ListenKey.Set(response.ListenKey)
	return nil
}

// RenewListenKey renews the ListenKey for user data stream
func (s *sBinance) RenewListenKey(apiKey string) error {
	req, err := http.NewRequest("PUT", apiBaseURL+listenKeyURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Println("关闭响应体错误：", err)
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return gerror.Newf("API error: %s", string(body))
	}

	return nil
}

// ConnectWebSocket safely connects to the WebSocket and updates conn
func (s *sBinance) ConnectWebSocket() error {
	// Close the existing connection if open
	if Conn != nil {
		err := Conn.Close()
		if err != nil {
			log.Println("Failed to close old connection:", err)
		}
	}

	// Create a new WebSocket connection
	wsURL := wsBaseURL + ListenKey.String()
	var err error
	Conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return gerror.Newf("failed to connect to WebSocket: %v", err)
	}

	log.Println("WebSocket connection established.")
	return nil
}
