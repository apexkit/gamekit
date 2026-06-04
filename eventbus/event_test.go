package eventbus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/apexkit/gamekit/eventbus/types"
	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

func TestNewEvent(t *testing.T) {
	payload := &types.WinEvent{
		GameType: "slot",
		Win:      100,
		Bet:      10,
	}
	evt := NewEvent("win", "api_server", payload)
	if evt == nil {
		t.Fatal("NewEvent returned nil")
	}
	if evt.Type != "win" {
		t.Errorf("Type = %q, want win", evt.Type)
	}
	if evt.Source != "api_server" {
		t.Errorf("Source = %q, want api_server", evt.Source)
	}
	if _, err := uuid.Parse(evt.EventID); err != nil {
		t.Errorf("EventID %q is not a valid UUID: %v", evt.EventID, err)
	}
	now := time.Now().Unix()
	if evt.Timestamp < now-2 || evt.Timestamp > now+2 {
		t.Errorf("Timestamp = %d, want near %d", evt.Timestamp, now)
	}

	var decoded types.WinEvent
	if err := evt.DecodePayload(&decoded); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if decoded.GameType != "slot" || decoded.Win != 100 || decoded.Bet != 10 {
		t.Errorf("decoded payload = %+v, want slot/100/10", decoded)
	}
}

func TestNewEvent_invalidPayload(t *testing.T) {
	ch := make(chan int)
	if evt := NewEvent("win", "api_server", ch); evt != nil {
		t.Errorf("NewEvent with unmarshalable payload should return nil, got %+v", evt)
	}
}

func TestEvent_EncodeDecodeEvent(t *testing.T) {
	original := &Event{
		EventID:   uuid.New().String(),
		Type:      "bet",
		Timestamp: 1700000000,
		Source:    "rtp_server",
		Metadata:  map[string]string{"trace_id": "abc-123"},
		Payload:   []byte(`{"amount":42}`),
	}

	data, err := original.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := DecodeEvent(data)
	if err != nil {
		t.Fatalf("DecodeEvent: %v", err)
	}
	if decoded.EventID != original.EventID {
		t.Errorf("EventID = %q, want %q", decoded.EventID, original.EventID)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, original.Type)
	}
	if decoded.Timestamp != original.Timestamp {
		t.Errorf("Timestamp = %d, want %d", decoded.Timestamp, original.Timestamp)
	}
	if decoded.Source != original.Source {
		t.Errorf("Source = %q, want %q", decoded.Source, original.Source)
	}
	if decoded.Metadata["trace_id"] != "abc-123" {
		t.Errorf("Metadata trace_id = %q, want abc-123", decoded.Metadata["trace_id"])
	}
	if string(decoded.Payload) != string(original.Payload) {
		t.Errorf("Payload = %s, want %s", decoded.Payload, original.Payload)
	}
}

func TestEvent_DecodePayload(t *testing.T) {
	type betPayload struct {
		Amount float64 `json:"amount"`
		Round  string  `json:"round"`
	}
	evt := NewEvent("bet", "api_server", &betPayload{Amount: 99.5, Round: "r-1"})
	if evt == nil {
		t.Fatal("NewEvent returned nil")
	}

	var out betPayload
	if err := evt.DecodePayload(&out); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if out.Amount != 99.5 || out.Round != "r-1" {
		t.Errorf("out = %+v, want amount=99.5 round=r-1", out)
	}
}

func TestEvent_DecodePayload_invalidJSON(t *testing.T) {
	evt := &Event{Payload: []byte("not json")}
	var out struct{}
	if err := evt.DecodePayload(&out); err == nil {
		t.Fatal("DecodePayload expected error for invalid JSON")
	}
}

func TestNewEvent_roundTripThroughEncode(t *testing.T) {
	in := NewEvent("win", "test", &types.WinEvent{
		PlayerId: "p1",
		Win:      50,
		GameType: "slot",
	})
	if in == nil {
		t.Fatal("NewEvent returned nil")
	}

	data, err := in.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	out, err := DecodeEvent(data)
	if err != nil {
		t.Fatalf("DecodeEvent: %v", err)
	}
	if out.Type != in.Type || out.Source != in.Source || out.EventID != in.EventID {
		t.Errorf("header mismatch: got type=%s source=%s id=%s", out.Type, out.Source, out.EventID)
	}

	var win types.WinEvent
	if err := out.DecodePayload(&win); err != nil {
		t.Fatalf("DecodePayload: %v", err)
	}
	if win.PlayerId != "p1" || win.Win != 50 || win.GameType != "slot" {
		t.Errorf("win = %+v, want p1/50/slot", win)
	}
}

func TestEvent_Encode_payloadIsBase64InJSON(t *testing.T) {
	evt := &Event{
		EventID: "test-id",
		Type:    "t",
		Payload: []byte(`{"k":"v"}`),
	}
	data, err := evt.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	// Go json 对 []byte 字段编码为 base64 字符串，不是嵌套 JSON 对象
	var payloadStr string
	if err := json.Unmarshal(raw["payload"], &payloadStr); err != nil {
		t.Fatalf("payload should be base64 string in outer JSON: %v", err)
	}
	if payloadStr == `{"k":"v"}` {
		t.Error("payload in outer JSON should be base64-encoded, not raw JSON object")
	}
}

func sampleWinEventPayload() *types.WinEvent {
	return &types.WinEvent{
		AppId:            "app-001",
		GameBrand:        "pg",
		GameType:         "slot",
		GameId:           "game-123",
		PlayerId:         "player-456",
		RoundId:          "round-789",
		Currency:         "USD",
		Bet:              10.5,
		Win:              100.0,
		BetTransactionId: "bet-tx-001",
		TransactionId:    "tx-001",
		Rtp:              "96.5",
	}
}

func sampleWinEventJSON() *Event {
	payload, _ := json.Marshal(sampleWinEventPayload())
	return &Event{
		EventID:   uuid.New().String(),
		Type:      "win",
		Timestamp: time.Now().Unix(),
		Source:    "api_server",
		Metadata:  map[string]string{"trace_id": "trace-abc-123"},
		Payload:   payload,
	}
}

func sampleWinEventMsgPack() (*Event, error) {
	payload, err := marshalMsgpack(sampleWinEventPayload())
	if err != nil {
		return nil, err
	}
	return &Event{
		EventID:   uuid.New().String(),
		Type:      "win",
		Timestamp: time.Now().Unix(),
		Source:    "api_server",
		Metadata:  map[string]string{"trace_id": "trace-abc-123"},
		Payload:   payload,
	}, nil
}

func encodeEventMsgpack(e *Event) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("json")
	if err := enc.Encode(e); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeEventMsgpack(data []byte) (*Event, error) {
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	dec.SetCustomStructTag("json")
	var evt Event
	if err := dec.Decode(&evt); err != nil {
		return nil, err
	}
	return &evt, nil
}

func marshalMsgpack(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.SetCustomStructTag("json")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshalMsgpack(data []byte, out interface{}) error {
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	dec.SetCustomStructTag("json")
	return dec.Decode(out)
}

const compareBenchIterations = 10000

func averageDuration(iterations int, fn func()) time.Duration {
	for i := 0; i < 100; i++ {
		fn()
	}
	start := time.Now()
	for i := 0; i < iterations; i++ {
		fn()
	}
	return time.Since(start) / time.Duration(iterations)
}

func speedCompareLine(jsonDur, msgpackDur time.Duration) string {
	if jsonDur == msgpackDur {
		return "—"
	}
	if jsonDur < msgpackDur {
		return fmt.Sprintf("JSON 快 %.1f%%", (float64(msgpackDur-jsonDur)/float64(msgpackDur))*100)
	}
	return fmt.Sprintf("MsgPack 快 %.1f%%", (float64(jsonDur-msgpackDur)/float64(jsonDur))*100)
}

func TestCompareJSONAndMsgPack(t *testing.T) {
	payload := sampleWinEventPayload()
	jsonEvt := sampleWinEventJSON()
	msgpackEvt, err := sampleWinEventMsgPack()
	if err != nil {
		t.Fatalf("sampleWinEventMsgPack: %v", err)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal payload: %v", err)
	}
	msgpackPayload, err := marshalMsgpack(payload)
	if err != nil {
		t.Fatalf("msgpack.Marshal payload: %v", err)
	}

	jsonEvent, err := jsonEvt.Encode()
	if err != nil {
		t.Fatalf("JSON Encode event: %v", err)
	}
	msgpackEvent, err := encodeEventMsgpack(msgpackEvt)
	if err != nil {
		t.Fatalf("MsgPack Encode event: %v", err)
	}

	fmt.Println("=== Payload 体积对比 ===")
	fmt.Printf("JSON:    %d bytes\n", len(jsonPayload))
	fmt.Printf("MsgPack: %d bytes\n", len(msgpackPayload))
	fmt.Printf("节省:    %.1f%%\n", (1-float64(len(msgpackPayload))/float64(len(jsonPayload)))*100)

	fmt.Println("=== 完整 Event 体积对比（payload 各自用对应格式） ===")
	fmt.Printf("JSON:    %d bytes\n", len(jsonEvent))
	fmt.Printf("MsgPack: %d bytes\n", len(msgpackEvent))
	fmt.Printf("节省:    %.1f%%\n", (1-float64(len(msgpackEvent))/float64(len(jsonEvent)))*100)
	fmt.Printf("JSON 示例:    %s\n", jsonEvent)
	fmt.Printf("MsgPack 示例: %x\n", msgpackEvent)

	jsonPayloadMarshal := averageDuration(compareBenchIterations, func() {
		_, _ = json.Marshal(payload)
	})
	msgpackPayloadMarshal := averageDuration(compareBenchIterations, func() {
		_, _ = marshalMsgpack(payload)
	})
	jsonPayloadUnmarshal := averageDuration(compareBenchIterations, func() {
		var out types.WinEvent
		_ = json.Unmarshal(jsonPayload, &out)
	})
	msgpackPayloadUnmarshal := averageDuration(compareBenchIterations, func() {
		var out types.WinEvent
		_ = unmarshalMsgpack(msgpackPayload, &out)
	})

	fmt.Println("=== Payload 速度对比 ===")
	fmt.Printf("Marshal  JSON:    %v/op\n", jsonPayloadMarshal)
	fmt.Printf("Marshal  MsgPack: %v/op  (%s)\n", msgpackPayloadMarshal, speedCompareLine(jsonPayloadMarshal, msgpackPayloadMarshal))
	fmt.Printf("Unmarshal JSON:    %v/op\n", jsonPayloadUnmarshal)
	fmt.Printf("Unmarshal MsgPack: %v/op  (%s)\n", msgpackPayloadUnmarshal, speedCompareLine(jsonPayloadUnmarshal, msgpackPayloadUnmarshal))

	jsonEventEncode := averageDuration(compareBenchIterations, func() {
		_, _ = jsonEvt.Encode()
	})
	msgpackEventEncode := averageDuration(compareBenchIterations, func() {
		_, _ = encodeEventMsgpack(msgpackEvt)
	})
	jsonEventDecode := averageDuration(compareBenchIterations, func() {
		_, _ = DecodeEvent(jsonEvent)
	})
	msgpackEventDecode := averageDuration(compareBenchIterations, func() {
		_, _ = decodeEventMsgpack(msgpackEvent)
	})

	fmt.Println("=== Event 速度对比 ===")
	fmt.Printf("Encode JSON:    %v/op\n", jsonEventEncode)
	fmt.Printf("Encode MsgPack: %v/op  (%s)\n", msgpackEventEncode, speedCompareLine(jsonEventEncode, msgpackEventEncode))
	fmt.Printf("Decode JSON:    %v/op\n", jsonEventDecode)
	fmt.Printf("Decode MsgPack: %v/op  (%s)\n", msgpackEventDecode, speedCompareLine(jsonEventDecode, msgpackEventDecode))

	if len(msgpackEvent) >= len(jsonEvent) {
		t.Errorf("MsgPack event size (%d) should be smaller than JSON (%d)", len(msgpackEvent), len(jsonEvent))
	}

	decodedJSON, err := DecodeEvent(jsonEvent)
	if err != nil {
		t.Fatalf("DecodeEvent JSON: %v", err)
	}
	decodedMsgpack, err := decodeEventMsgpack(msgpackEvent)
	if err != nil {
		t.Fatalf("DecodeEvent MsgPack: %v", err)
	}

	for name, pair := range map[string][2]*Event{
		"JSON":    {jsonEvt, decodedJSON},
		"MsgPack": {msgpackEvt, decodedMsgpack},
	} {
		if pair[1].EventID != pair[0].EventID || pair[1].Type != pair[0].Type || pair[1].Source != pair[0].Source {
			t.Errorf("%s round-trip header mismatch: got id=%s type=%s source=%s",
				name, pair[1].EventID, pair[1].Type, pair[1].Source)
		}
	}

	var winFromJSON, winFromMsgpack types.WinEvent
	if err := decodedJSON.DecodePayload(&winFromJSON); err != nil {
		t.Fatalf("DecodePayload JSON: %v", err)
	}
	if err := unmarshalMsgpack(decodedMsgpack.Payload, &winFromMsgpack); err != nil {
		t.Fatalf("DecodePayload MsgPack: %v", err)
	}
	if winFromJSON != winFromMsgpack {
		t.Errorf("payload mismatch: json=%+v msgpack=%+v", winFromJSON, winFromMsgpack)
	}
}

func BenchmarkEventEncodeJSON(b *testing.B) {
	evt := sampleWinEventJSON()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := evt.Encode(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventEncodeMsgPack(b *testing.B) {
	evt, err := sampleWinEventMsgPack()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := encodeEventMsgpack(evt); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventDecodeJSON(b *testing.B) {
	evt := sampleWinEventJSON()
	data, err := evt.Encode()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DecodeEvent(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventDecodeMsgPack(b *testing.B) {
	evt, err := sampleWinEventMsgPack()
	if err != nil {
		b.Fatal(err)
	}
	data, err := encodeEventMsgpack(evt)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := decodeEventMsgpack(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPayloadMarshalJSON(b *testing.B) {
	payload := sampleWinEventPayload()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(payload); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPayloadMarshalMsgPack(b *testing.B) {
	payload := sampleWinEventPayload()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := marshalMsgpack(payload); err != nil {
			b.Fatal(err)
		}
	}
}
