package client

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	pb "wails/backend/chatProto"
)

var (
	serverAddr = flag.String("addr", "3.38.170.112:50051", "the address to connect to")
)

type Client struct {
	conn           *grpc.ClientConn                           // 연결 정보
	client         pb.ChatServiceClient                       // 클라이언트
	streams        map[string]pb.ChatService_ChatStreamClient // 클라이언트 스트림(키: 채팅방 ID, 값: 스트림)
	ctx            context.Context
	messageChans   map[string]chan JSMessage
	messageHistory map[string][]JSMessage
	wailsCtx       context.Context // wails 리액트 전달을 위한 컨텍스트
}

// 클라이언트 요청에 대한 마샬링을 위한 구조체
type JSMessage struct {
	Channel   string `json:"channel"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// gRPC 서버에 연결 및 새로운 클라이언트 생성
func NewClient() (*Client, error) {
	// gRPC 서버에 연결
	// TODO: TLS 인증서 설정 필요, 현재는 설정 안함
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("에러 발생: %v", err)
		return nil, err
	}

	// 클라이언트 생성
	client := pb.NewChatServiceClient(conn)
	log.Printf("클라이언트 생성 완료")

	ctx := context.Background()

	return &Client{
		conn:           conn,
		client:         client,
		streams:        make(map[string]pb.ChatService_ChatStreamClient),
		ctx:            ctx,
		messageChans:   make(map[string]chan JSMessage, 10),
		messageHistory: make(map[string][]JSMessage),
		wailsCtx:       nil,
	}, nil
}

func (c *Client) SetWailsContext(ctx context.Context) {
	c.wailsCtx = ctx
}

// 사용자 구분을 위한 uuid 발급
func (c *Client) SetUserId() {
	md, ok := metadata.FromOutgoingContext(c.ctx)
	var uid string

	if ok { // 메타데이터 있음
		val := md.Get("user_uuid")
		if len(val) > 0 { // 패닉 방지
			uid = val[0]
			log.Printf("%s가 이미 설정되어 있음", uid)
		}
	}

	uid = uuid.NewString()
	c.ctx = metadata.AppendToOutgoingContext(c.ctx, "user_uuid", uid)
	log.Printf("%s 설정 완료", uid)
}

func (c *Client) GetUserId() string {
	md, ok := metadata.FromOutgoingContext(c.ctx)
	if ok {
		vals := md.Get("user_uuid")
		if len(vals) > 0 { // 패닉 방지
			return vals[0]
		}
	}

	return ""
}

// 양방향 스트리밍 채팅 시작
func (c *Client) StartChat(chanId string) error {
	if c.streams[chanId] != nil {
		log.Printf("스트림이 이미 열려 있음")
		return nil
	}

	stream, err := c.client.ChatStream(c.ctx)
	if err != nil {
		log.Printf("스트림 생성 실패: %v", err)
		return err
	}

	c.streams[chanId] = stream
	c.messageChans[chanId] = make(chan JSMessage, 10)

	log.Printf("스트림 생성 완료")

	go c.ReceiveMessages(chanId)
	go c.SendMessages(chanId)

	return nil
}

// 클라이언트에서 고루틴 채널로 메시지 송신
func (c *Client) SendChatMessage(msg JSMessage) error {
	chanId := msg.Channel
	messageChan := c.messageChans[chanId]

	if messageChan == nil {
		return fmt.Errorf("메시지 채널이 없음")
	}

	select {
	case messageChan <- msg:
		return nil
	default:
		return fmt.Errorf("메시지 채널이 포화 또는 사용 불가 상태")
	}
}

// 서버로 메시지 송신
func (c *Client) SendMessages(chanId string) {
	for jsMsg := range c.messageChans[chanId] {
		// JSMessage -> pb.ChatMessage로 변환
		// time 변환
		parsedTime, err := time.Parse(time.RFC3339Nano, jsMsg.Timestamp)
		if err != nil { // 제로 타임 방지
			parsedTime = time.Now()
		}

		// pb.ChatMessage 구조체 생성 및 포인터 할당
		pbMsg := &pb.ChatMessage{
			Channel:   jsMsg.Channel,
			Sender:    jsMsg.Sender,
			Content:   jsMsg.Content,
			Timestamp: timestamppb.New(parsedTime),
		}

		// 스트림이 nil이면 그냥 무시
		if c.streams[chanId] == nil {
			log.Printf("스트림 없음, 메시지 무시: %s", pbMsg.Content)
			continue
		}

		if err := c.streams[chanId].Send(pbMsg); err != nil {
			log.Printf("전송 실패: %v", err)
			continue
		}
		log.Printf("전송 성공: %s", pbMsg.Content)
	}

	log.Printf("SendMessages 종료: 채널 닫힘")
}

func (c *Client) AddMessageHistory(msg JSMessage) {
	chanId := msg.Channel
	if c.messageHistory[chanId] == nil {
		c.messageHistory[chanId] = []JSMessage{}
	}

	c.messageHistory[chanId] = append(c.messageHistory[chanId], msg)
	if len(c.messageHistory[chanId]) > 50 {
		c.messageHistory[chanId] = c.messageHistory[chanId][1:]
	}
}

// 채팅방 메시지 기록 반환 함수 추가
func (c *Client) GetMessageHistory(chanId string) []JSMessage {
	return c.messageHistory[chanId]
}

// 서버로부터 메시지 수신
func (c *Client) ReceiveMessages(chanId string) error {
	for {
		if c.streams[chanId] == nil {
			log.Printf("ReceiveMessages: 스트림이 닫힘")
			break
		}

		// 서버로부터 수신 대기
		in, err := c.streams[chanId].Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf("ReceiveMessages: 스트림 정상 종료 (EOF)")
			} else {
				log.Printf("ReceiveMessages: 스트림 수신 오류: %v", err)
			}
			break
		}

		log.Printf("[Go] gRPC에서 메시지 수신: %+v", in)

		// wails 전달을 위해 JSMessage로 변환
		jsMsg := JSMessage{
			Channel:   in.GetChannel(),
			Sender:    in.GetSender(),
			Content:   in.GetContent(),
			Timestamp: in.GetTimestamp().AsTime().Format(time.RFC3339Nano),
		}
		c.AddMessageHistory(jsMsg)

		if c.wailsCtx != nil {
			log.Printf("wails로 메시지 전달: %+v", jsMsg)
			runtime.EventsEmit(c.wailsCtx, "newMessage", jsMsg)
		}
	}
	log.Printf("ReceiveMessages 고루틴 종료.")
	return nil
}

func (c *Client) GetChannelList() []string {
	result, err := c.client.ListChannels(c.ctx, &pb.ListChannelsRequest{})
	if err != nil {
		log.Printf("서버 요청 실패: %v", err)
		return []string{} // 빈 slice 반환
	}

	if result == nil {
		log.Printf("서버 응답이 nil")
		return []string{}
	}

	return result.ChannelIds
}

// 단일 스트림 연결 종료
func (c *Client) CloseChat(chanId string) {
	// 스트림 Send 종료
	if c.streams[chanId] != nil {
		if err := c.streams[chanId].CloseSend(); err != nil {
			log.Printf("Close: 스트림 Send 종료 실패: %v", err)
		}
		c.streams[chanId] = nil
	}

	c.messageHistory[chanId] = []JSMessage{}
	log.Printf("채팅방 나가기 완료")
}

// 클라이언트 연결 종료
func (c *Client) Close() {
	// 메시지 채널 닫기
	if c.messageChans != nil {
		for _, messageChan := range c.messageChans {
			close(messageChan)
			messageChan = nil
		}
	}

	// 모든 스트림 Send 종료
	if c.streams != nil {
		for _, stream := range c.streams {
			if err := stream.CloseSend(); err != nil {
				log.Printf("Close: 스트림 Send 종료 실패: %v", err)
			}
			stream = nil
		}
	}

	// gRPC 연결 종료
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Close: gRPC 연결 종료 실패: %v", err)
		}
		c.conn = nil
	}

	log.Printf("클라이언트 연결 종료 절차 완료")
}
