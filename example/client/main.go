package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	head     byte   = 0x7E
	serverIP string = "localhost:12336"
)

//Packet .
type Packet struct {
	head    byte
	msgID   uint16
	cmdType byte
	pData   []byte
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()
	for i := 10000; i < 10001; i++ {
		temp := i
		connection, err := CreateTCPConn(serverIP)
		if err != nil {
			log.Println(err)
			continue
		}
		go func(conn net.Conn, num int) {
			defer recover()
			var msgID uint16 = 100
			cmdType := byte(0x00)
			start := "AAAAAAAAAA"
			end := "BBBBBBBBB"
			pData := []byte(start + strconv.Itoa(num) + end)
			SendTCPCMD(connection, Packet{head, msgID, cmdType, pData})
			ReceivTCPCMD(connection, 0)
			cmdType = byte(0x01)
			count := 0
			for {
				// if count == 10*(temp-10000) {
				// 	fmt.Println("exit")
				// 	// conn.Close()
				// 	return
				// }
				pData = []byte(fmt.Sprintf("Client:%d,心跳:第%d次心跳", num, count))
				SendTCPCMD(conn, Packet{head, msgID, cmdType, pData})
				ReceivTCPCMD(connection, count)
				time.Sleep(1 * time.Second)
				count++
			}
		}(connection, temp)
	}
	<-context.Background().Done()
}

//CreateTCPConn 创建一个TCP连接
func CreateTCPConn(serverIP string) (net.Conn, error) {
	hawkServer, err := net.ResolveTCPAddr("tcp", serverIP)
	if err != nil {
		return nil, err
	}
	connection, err := net.DialTCP("tcp", nil, hawkServer)
	if err != nil {
		return nil, err
	}
	return connection, err
}

//GetSerializePacket Packet包序列化
func GetSerializePacket(packet Packet) []byte {
	willRecvLen := uint16(1 + 2 + 2 + 1 + len(packet.pData) + 1) // 帧长    msgid+命令类型+datainfo+校验和
	slice := make([]byte, 0)
	slice = append(slice, packet.head)                    //帧头		1
	slice = append(slice, UInt16ToBytes(willRecvLen)...)  //帧长	2
	slice = append(slice, UInt16ToBytes(packet.msgID)...) //消息编号	2
	slice = append(slice, packet.cmdType)                 //命令类型	1
	slice = append(slice, packet.pData...)                //发送文本	N
	check := GetChecksum(slice)
	slice = append(slice, byte(check)) //校验码	1
	return slice
}

//GetDeserializationPacket Packet包反序列化
func GetDeserializationPacket(data []byte) (Packet, error) {
	//  检验和的值要和  整包的长度一致，包括 校验和 字节
	//  帧长 告诉服务端之后要接受多少个字节
	head := data[0]
	msgID := BytesToUInt16(data[3:5])
	cmdType := data[5]
	pData := data[6 : len(data)-1]
	check := data[len(data)-1]
	if GetChecksum(data[0:len(data)-1]) == check {
		return Packet{head, msgID, cmdType, pData}, nil
	}
	return Packet{}, errors.New("检验不通过")
}

//SendTCPCMD 发送命令
func SendTCPCMD(connection net.Conn, packet Packet) error {
	slice := GetSerializePacket(packet)
	_, err := connection.Write(slice)
	if err != nil {
		log.Println("发送异常：", err.Error())
		return err
	}
	return nil
}

//ReceivTCPCMD .
func ReceivTCPCMD(connection net.Conn, count int) error {
	slice := make([]byte, 5120)
	len, err := connection.Read(slice)
	if len == 0 {
		return nil
	}
	p, err := GetDeserializationPacket(slice[:len])
	{
		fmt.Println(string(p.pData))
	}
	return err
}

// GetChecksum 校验和
func GetChecksum(raw []byte) uint8 {
	var sum int64
	for i := 0; i < len(raw); i++ {

		sum += int64(raw[i])
	}
	return ^uint8(sum) + 1
}

//UInt16ToBytes .
func UInt16ToBytes(n uint16) []byte {
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, n)
	return result
}

//BytesToUInt16 .
func BytesToUInt16(array []byte) uint16 {
	result := binary.BigEndian.Uint16(array)
	return result
}
