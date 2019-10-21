package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

// IoTButtonEvent : AWS IoTボタン押下時のイベントデータ
type IoTButtonEvent struct {
	DeviceEvent struct {
		ButtonClicked struct {
			ClickType    string    `json:"clickType"`
			ReportedTime time.Time `json:"reportedTime"`
		} `json:"buttonClicked"`
	} `json:"deviceEvent"`
	DeviceInfo struct {
		Type          string  `json:"type"`
		DeviceID      string  `json:"deviceId"`
		RemainingLife float64 `json:"remainingLife"`
	} `json:"deviceInfo"`
	PlacementInfo struct {
		ProjectName   string `json:"projectName"`
		PlacementName string `json:"placementName"`
	} `json:"placementInfo"`
}

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(event IoTButtonEvent) {

	// 環境変数からイベント種別に対応した施錠/解錠動作コマンドを取得
	fmt.Printf("IoTButtonEvent: %v\n", event)
	clickType := event.DeviceEvent.ButtonClicked.ClickType
	cmd := os.Getenv(clickType)
	fmt.Printf("clickType: [%s]\n", clickType)
	fmt.Printf("SendCommand: [%s]\n", cmd)

	// 動作指定がなければ終了
	switch cmd {
	case "lock": // 処理継続
	case "unlock": // 処理継続
	case "sync": // 処理継続
	default: // それ以外は終了
		return
	}

	// 2つの錠に対して同時に(並列して)コマンドを送る
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		postCommand(os.Getenv("DEVICE1"), cmd)
	}()
	go func() {
		defer wg.Done()
		postCommand(os.Getenv("DEVICE2"), cmd)
	}()
	wg.Wait()
}

func postCommand(deviceID, command string) {

	jsonStr := `{"command":"` + command + `"}`
	//jsonStr := fmt.Sprintf(`{"channel":"%s", "username":"Lambda", "text":"[%s] %s"}`, os.Getenv("CHANNEL"), command, deviceID) // Slack通知
	req, err := http.NewRequest(
		"POST",
		"https://api.candyhouse.co/public/sesame/"+deviceID,
		//os.Getenv("SLACK"), // Sesame使わず事前検証のためのSlack通知
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		fmt.Print(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", os.Getenv("APIKEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Printf("Response from [%s] %v\n", deviceID, resp)
	defer resp.Body.Close()
}
