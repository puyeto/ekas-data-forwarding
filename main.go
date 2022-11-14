package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

// HandleRequest Handles incoming requests.
func HandleRequest(conn net.Conn) {
	// set 2 minutes timeout
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	defer conn.Close()

	var byteSize = 70
	byteData := make([]byte, 700)

	for {
		reqLen, err := conn.Read(byteData)
		if err != nil {
			if err != io.EOF {
				fmt.Println("End of file error:", err)
			}
			fmt.Println("Error reading:", err.Error(), reqLen)
			return
		}

		// return Response
		result := "Received byte size = " + strconv.Itoa(reqLen) + "\n"
		conn.Write([]byte(string(result)))

		if reqLen == 0 {
			return // connection already closed by client
		}

		if reqLen > 0 {
			byteRead := bytes.NewReader(byteData)

			for i := 0; i < (reqLen / byteSize); i++ {

				byteRead.Seek(int64((byteSize * i)), 0)

				mb := make([]byte, byteSize)
				n1, _ := byteRead.Read(mb)

				go processRequest(conn, mb, n1)
			}

		}
	}
}

func main() {
	log.SetPrefix("tcpforward: ")

	listener, err := net.Listen("tcp", ":6033")
	log.Println("Listening on port 6033")
	if err != nil {
		log.Printf("Failed to setup listener: %v", err)
	}

	defer listener.Close()
	rand.Seed(time.Now().Unix())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("ERROR: failed to accept listener: %v", err)
		}
		log.Printf("Accepted connection from %v\n", conn.RemoteAddr().String())
		go HandleRequest(conn)
	}
}

func hasBit(n int, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func readInt32(data []byte) (ret int32) {
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &ret)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}

	return ret
}

func readNextBytes(conn net.Conn, number int) (int, []byte) {
	bytes := make([]byte, number)

	reqLen, err := conn.Read(bytes)
	if err != nil {
		if err != io.EOF {
			fmt.Println("End of file error:", err)
		}
		fmt.Println("Error reading:", err.Error(), reqLen)
	}

	return reqLen, bytes
}

func processRequest(conn net.Conn, b []byte, byteLen int) {
	var deviceData DeviceData

	if byteLen != 70 {
		fmt.Println("Invalid Byte Length = ", byteLen)
		return
	}

	byteReader := bytes.NewReader(b)

	scode := make([]byte, 4)
	byteReader.Read(scode)
	deviceData.SystemCode = string(scode)
	if deviceData.SystemCode != "MCPG" {
		fmt.Println("data not valid", deviceData.SystemCode)
		fmt.Println("device data", deviceData)
		return
	}

	byteReader.Seek(5, 0)
	did := make([]byte, 4)
	byteReader.Read(did)
	deviceData.DeviceID = binary.LittleEndian.Uint32(did)
	if deviceData.DeviceID == 0 {
		return
	}

	// fmt.Println(deviceData.DeviceID, time.Now(), " data received")

	// Transmission Reason – 1 byte
	byteReader.Seek(18, 0)
	reason := make([]byte, 1)
	byteReader.Read(reason)
	deviceData.TransmissionReason = int(reason[0])

	// Transmission Reason Specific data – 1 byte
	trsd := 0
	if deviceData.TransmissionReason == 255 {
		byteReader.Seek(17, 0)
		specific := make([]byte, 1)
		byteReader.Read(specific)

		var a = int(specific[0])
		// Failsafe
		failsafe := hasBit(a, 1)
		deviceData.Failsafe = failsafe
		// main power disconnected
		disconnect := hasBit(a, 2)
		deviceData.Disconnect = disconnect
		trsd = int(a)
	}
	deviceData.TransmissionReasonSpecificData = trsd

	// Number of satellites used (from GPS) – 1 byte
	byteReader.Seek(43, 0)
	satellites := make([]byte, 1)
	byteReader.Read(satellites)
	deviceData.NoOfSatellitesUsed = int(satellites[0])

	// Longitude – 4 bytes
	byteReader.Seek(44, 0)
	long := make([]byte, 4)
	byteReader.Read(long)
	deviceData.Longitude = readInt32(long)

	//  Latitude – 4 bytes
	byteReader.Seek(48, 0)
	lat := make([]byte, 4)
	byteReader.Read(lat)
	deviceData.Latitude = readInt32(lat)

	// Altitude
	byteReader.Seek(52, 0)
	alt := make([]byte, 4)
	byteReader.Read(alt)
	deviceData.Altitude = readInt32(alt)

	// Ground speed – 4 bytes
	byteReader.Seek(56, 0)
	gspeed := make([]byte, 4)
	byteReader.Read(gspeed)
	deviceData.GroundSpeed = binary.LittleEndian.Uint32(gspeed)

	// Speed direction – 2 bytes
	byteReader.Seek(60, 0)
	speedd := make([]byte, 2)
	byteReader.Read(speedd)
	deviceData.SpeedDirection = int(binary.LittleEndian.Uint16(speedd))

	// UTC time – 3 bytes (hours, minutes, seconds)
	byteReader.Seek(62, 0)
	sec := make([]byte, 1)
	byteReader.Read(sec)
	deviceData.UTCTimeSeconds = int(sec[0])

	byteReader.Seek(63, 0)
	min := make([]byte, 1)
	byteReader.Read(min)
	deviceData.UTCTimeMinutes = int(min[0])

	byteReader.Seek(64, 0)
	hrs := make([]byte, 1)
	byteReader.Read(hrs)
	deviceData.UTCTimeHours = int(hrs[0])

	// UTC date – 4 bytes (day, month, year)
	byteReader.Seek(65, 0)
	day := make([]byte, 1)
	byteReader.Read(day)
	deviceData.UTCTimeDay = int(day[0])

	byteReader.Seek(66, 0)
	mon := make([]byte, 1)
	byteReader.Read(mon)
	deviceData.UTCTimeMonth = int(mon[0])

	byteReader.Seek(67, 0)
	yr := make([]byte, 2)
	byteReader.Read(yr)
	deviceData.UTCTimeYear = int(binary.LittleEndian.Uint16(yr))

	deviceData.DateTime = time.Date(deviceData.UTCTimeYear, time.Month(deviceData.UTCTimeMonth), deviceData.UTCTimeDay, deviceData.UTCTimeHours, deviceData.UTCTimeMinutes, deviceData.UTCTimeSeconds, 0, time.UTC)
	deviceData.DateTimeStamp = deviceData.DateTime.Unix()

	// if checkIdleState(deviceData) != "idle3" {
	// clientJobs <- models.ClientJob{deviceData, conn}
	//}

	fmt.Println(deviceData)
	url := "http://equscabanus.com:6055?id=" + strconv.Itoa(int(deviceData.DeviceID))
	url += "&lat=" + strconv.Itoa(int(deviceData.Latitude/10000000)) + "&lon=" + strconv.Itoa(int(deviceData.Longitude/10000000))
	url += "&timestamp=" + strconv.Itoa(int(deviceData.DateTimeStamp)) + "&altitude=" + strconv.Itoa(int(deviceData.Altitude))
	url += "&speed=5"
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s", err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
		}
		fmt.Printf("%s\n", string(contents))
	}

}
