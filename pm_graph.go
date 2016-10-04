package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	fileName string = "pm.json"
	longForm        = "2006-01-02 15:04:05.999999999 -0700 MST"
)

type PMData struct {
	TimeStamp string
	Value     float64
}

type Object struct {
	Class    string
	Data     []PMData
	ModuleId uint8
	NwIntfId uint8
	Resource string
	Type     string
}

type PMJson struct {
	Object   Object
	ObjectId string
}

//func parsePMJson(pmJson *PMJson) error {
func parsePMJson(pmJson *PMJson, bytes []byte) error {
	/*
		bytes, err := ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
	*/
	return json.Unmarshal(bytes, pmJson)
}

type PMDataPoint struct {
	TimeStamp int64
	Value     float64
}

func getPMDataPoints(pmJson *PMJson) []PMDataPoint {
	var pmDataPoints []PMDataPoint
	var dataPoint PMDataPoint
	for _, data := range pmJson.Object.Data {
		dataPoint.Value = data.Value
		ti, _ := time.Parse(longForm, data.TimeStamp)
		dataPoint.TimeStamp = ti.Unix()
		pmDataPoints = append(pmDataPoints, dataPoint)
	}
	return pmDataPoints
}

func Points(dataPoint []PMDataPoint) plotter.XYs {
	pts := make(plotter.XYs, len(dataPoint))
	i := 0
	for _, data := range dataPoint {
		pts[i].X = float64(data.TimeStamp - dataPoint[len(dataPoint)-1].TimeStamp)
		pts[i].Y = data.Value
		i++
	}
	return pts
}

//func plotGraph(dataPoints []PMDataPoint, pmJson *PMJson, minTime int64) error {
func plotGraph(dataPoints []PMDataPoint, pmJson *PMJson, fileName string) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = fmt.Sprintf("ModuleId: %d, NwIntfId: %d, Resource:%s, Type:%s, Class:%s", pmJson.Object.ModuleId, pmJson.Object.NwIntfId, pmJson.Object.Resource, pmJson.Object.Type, pmJson.Object.Class)
	//p.X.Label.Text = fmt.Sprintf("Time (In Second, Start= %d)", minTime)
	p.X.Label.Text = fmt.Sprintf("Time (second)")
	p.Y.Label.Text = fmt.Sprintf("Resource:%s, Type: %s In(dB)", pmJson.Object.Resource, pmJson.Object.Type)
	err = plotutil.AddLinePoints(p, "First", Points(dataPoints))
	if err != nil {
		return err
	}

	if err := p.Save(10*vg.Inch, 10*vg.Inch, fileName); err != nil {
		return err
	}
	return err
}

type PMQueryStruct struct {
	ModuleId int
	NwIntfId int
	Resource string
	Type     string
	Class    string
}

func getPMData(ipAddr, port, moduleId, nwIntfId, resource, Type, class string) []byte {
	url := "http://" + ipAddr + ":" + port + "/public/v1/state/DWDMModuleNwIntfPM"
	modId, _ := strconv.Atoi(moduleId)
	nwId, _ := strconv.Atoi(nwIntfId)
	data := PMQueryStruct{
		ModuleId: modId,
		NwIntfId: nwId,
		Resource: resource,
		Type:     Type,
		Class:    class,
	}

	jsonStr, _ := json.Marshal(data)
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	return SendHttpCmd(req)

}

func SendHttpCmd(req *http.Request) []byte {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))
	return body
}

func main() {
	ipAddr := flag.String("IP", "localhost", "Ip Address")
	port := flag.String("Port", "8080", "port number")
	moduleId := flag.String("ModuleId", "0", "Module Id")
	nwIntfId := flag.String("NwIntfId", "0", "Network Interface Id")
	resource := flag.String("Resource", "BER", "Resource Name")
	Type := flag.String("Type", "Current", "Current/Min/Max/Avg")
	class := flag.String("Class", "Class-A", "Class-A/Class-B/Class-C")
	outputFileName := flag.String("OutputFile", "PM.png", ".png file")
	flag.Parse()
	jsonData := getPMData(*ipAddr, *port, *moduleId, *nwIntfId, *resource, *Type, *class)
	pmJson := new(PMJson)
	//err := parsePMJson(pmJson)
	err := parsePMJson(pmJson, jsonData)
	if err != nil {
		fmt.Println("Error Parsing the PM Json", err)
		return
	}
	pmDataPoints := getPMDataPoints(pmJson)
	err = plotGraph(pmDataPoints, pmJson, *outputFileName)
	if err != nil {
		fmt.Println("Error while plotting graph", err)
		return
	}
}
