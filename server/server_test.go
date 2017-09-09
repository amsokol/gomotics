package server_test

// TODO: review the http testing
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mch1307/gomotics/db"
	"github.com/mch1307/gomotics/log"
	. "github.com/mch1307/gomotics/server"
	"github.com/mch1307/gomotics/testutil"
	"github.com/mch1307/gomotics/types"
)

var baseUrl string

//const healthMsg = `{"alive":true}`

func init() {
	fmt.Println("starting server test")
	baseUrl = "http://" + testutil.ConnectHost + ":8081"
	testutil.InitStubNHC()
	time.Sleep(time.Second * 1)
}

func TestHealth(t *testing.T) {
	req, err := http.NewRequest("GET", baseUrl+"/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Health)
	handler.ServeHTTP(rr, req)
	if rr.Body.String() != HealthMsg {
		t.Errorf("health test failed: got %v, expect: %v", rr.Body.String(), HealthMsg)
	}

}

// TODO: add more test cases (test non existing item)
func Test_getNhcItem(t *testing.T) {
	req, err := http.NewRequest("GET", baseUrl+"/api/v1/nhc/99", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetNhcItem)
	handler.ServeHTTP(rr, req)
	expected := "light"
	var res types.Item
	json.Unmarshal(rr.Body.Bytes(), &res)
	if res.Name != expected {
		t.Errorf("getNhcItem failed: got %v, expect: %v", res, expected)
	}
}

func Test_getNhcItems(t *testing.T) {
	req, err := http.NewRequest("GET", baseUrl+"/api/v1/nhc", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetNhcItems)
	handler.ServeHTTP(rr, req)
	var found bool
	expected := "light"
	var res []types.Item
	json.Unmarshal(rr.Body.Bytes(), &res)
	for _, val := range res {
		if val.ID == 0 {
			if val.Name == expected {
				found = true
			}
		}
	}
	if !found {
		t.Error("GetNhcItems failed, expected light record not found")
	}
}

func Test_nhcCmd(t *testing.T) {
	expected := "Success"
	url := baseUrl + "/api/v1/nhc/1/100"
	hCli := http.Client{
		Timeout: time.Second * 2,
	}
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	//	req.Header.Set("User-Agent", "Test_nhcCmd")
	rsp, getErr := hCli.Do(req)
	if getErr != nil {
		fmt.Println(err)
	}
	got, readErr := ioutil.ReadAll(rsp.Body)
	if readErr != nil {
		fmt.Println("Read err: ", readErr)
	}
	//defer rsp.Body.Close()
	if string(got) != expected {
		t.Errorf("Test_nhcCmd failed, expecting %v, got %v", expected, string(got))
	}

}

func TestGetNhcInfo(t *testing.T) {
	expected := "1.10.0.34209"
	url := baseUrl + "/api/v1/nhc/info"
	hCli := http.Client{
		Timeout: time.Second * 2,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", "Test_nhcCmd")
	rsp, getErr := hCli.Do(req)
	if getErr != nil {
		fmt.Println(err)
	}
	got, readErr := ioutil.ReadAll(rsp.Body)
	if readErr != nil {
		fmt.Println("Read err: ", readErr)
	}
	var res types.NHCSystemInfo
	json.Unmarshal(got, &res)
	//defer rsp.Body.Close()
	if res.Swversion != expected {
		t.Errorf("TestGetNhcInfo failed, expecting %v, got %v", expected, res.Swversion)
	}

}

var origin = "http://localhost/"
var url = "ws://localhost:8081/events"

func wsDial(url string) (wsConn *websocket.Conn, ok bool, err error) {
	webS, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("error connecting ws", err)
		return webS, false, err
	}
	//fmt.Println("websocket connect ok")
	return webS, true, nil
}

func Test_tWS(t *testing.T) {
	retry := 0
	ok := false
	ctl := 0
	var err error
	var wsConn *websocket.Conn
	tests := []struct {
		name       string
		id         int
		exName     string
		exLocation string
		exState    int
	}{
		{"action0", 3, "light", "Living Room", 0},
		{"action1", 1, "power switch", "Kitchen", 100},
	}
	//fmt.Println("# tests: ", len(tests))
	var msg types.Item
	time.Sleep(time.Second * 2)
	if wsConn, ok, err = wsDial(url); !ok {
		if retry < 11 {
			retry++
			fmt.Println("Retrying websocket connect due to error: ", err)
			fmt.Println("Attempt # ", retry)
			time.Sleep(time.Second * 1)
			Test_tWS(t)
		} else {
			fmt.Println("Could not connect after 10 attempts, err: ", err)
			return
		}
	}

	//time.Sleep(time.Second * 2)
	go func() {
		//defer ws.Close()
		//var tmp = make([]byte, 512)
		for {
			_, tmp, err := wsConn.ReadMessage()
			if err != nil {
				log.Error("read:", err)
				return
			}
			log.Error(string(tmp))
			err = json.Unmarshal(bytes.TrimSpace(tmp), &msg)
			if err != nil {
				log.Error("err parsing: ", err)
				log.Error(string(tmp))
			}
			//fmt.Println("ws reads ", msg)

		}
	}()

	time.Sleep(time.Second * 3)
	for _, tt := range tests {
		fmt.Println("start test ", tt.name)
		//ws.WriteMessage(websocket.PingMessage, nil)
		/* 		cmd := testutil.MyCmd
		   		cmd.ID = tt.id
		   		cmd.Value = tt.exState */
		//fmt.Println(cmd)
		time.Sleep(time.Millisecond * 500)
		var evt types.Event
		evt.ID = tt.id
		evt.Value = tt.exState
		db.ProcessEvent(evt)
		time.Sleep(time.Millisecond * 500)

		//fmt.Println("msg ", msg.ID)
		if msg.ID != tt.id || (msg.State != tt.exState) {
			t.Error("test failed  ", tt.name, tt.id, msg.ID, tt.exName, msg.Name, tt.exState, msg.State)
		}
		ctl++
	}
	defer wsConn.Close()
	//fmt.Println("tests ok: ", ctl)
}

func stubNHCTCP() {
	// listen to incoming tcp connections
	l, err := net.Listen("tcp", "0.0.0.0:8000")
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()
	_, err = l.Accept()
	if err != nil {
		fmt.Println(err)
	}
}

func stubNHCUDP() {
	// listen to incoming udp packets
	fmt.Println("starting UDP stub")
	pc, err := net.ListenPacket("udp", "0.0.0.0:10000")
	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	//simple read
	buffer := make([]byte, 1024)
	var addr net.Addr
	_, addr, _ = pc.ReadFrom(buffer)

	//simple write
	pc.WriteTo([]byte("NHC Stub"), addr)
}

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func TestDiscover(t *testing.T) {

	tests := []struct {
		name string
		want net.IP
	}{
		{"no nhc on LAN", nil},
		//{"stub nhc", getOutboundIP()},
	}
	portCheckIteration := 0
	for _, tt := range tests {
		fmt.Println("starting test ", tt.name)
		if tt.want != nil {
			go stubNHCUDP()
			go stubNHCTCP()
		}
		t.Run(tt.name, func(t *testing.T) {
		GotoTestPort:
			if testutil.IsTCPPortAvailable(18043) {
				if got := Discover(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Discover() = %v, want %v", got, tt.want)
				}
			} else {
				portCheckIteration++
				if portCheckIteration < 21 {
					fmt.Printf("UDP 18043 busy, %v retry", portCheckIteration)
					time.Sleep(time.Millisecond * 500)
					goto GotoTestPort
				} else {
					t.Error("Discover failed to get UDP port 18043, test failed")
				}

			}
		})
	}
}
