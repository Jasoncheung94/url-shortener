package cache

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		key           string
		value         interface{}
		ttl           time.Duration
		mockBehavior  func(mock redismock.ClientMock)
		expectedError error
	}{
		{
			name:  "Set successful",
			key:   "testKey",
			value: "testValue",
			ttl:   time.Second * 10,
			mockBehavior: func(mock redismock.ClientMock) {
				serializedValue, _ := json.Marshal("testValue")
				mock.ExpectSet("testKey", serializedValue, time.Second*10).SetVal("OK")
			},
			expectedError: nil,
		},
		{
			name:  "Set error from Redis",
			key:   "testKey",
			value: "testValue",
			ttl:   time.Second * 10,
			mockBehavior: func(mock redismock.ClientMock) {
				serializedValue, _ := json.Marshal("testValue")
				mock.ExpectSet("testKey", serializedValue, time.Second*10).SetErr(errors.New("test redis error"))
			},
			expectedError: errors.New("test redis error"),
		},
		{
			name:  "Marshal error",
			key:   "testKey",
			value: make(chan int), // Invalid for JSON
			ttl:   time.Second * 10,
			mockBehavior: func(mock redismock.ClientMock) {
				// No Redis interaction
			},
			expectedError: errors.New("json: unsupported type: chan int"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock := redismock.NewClientMock()
			cache := NewRedis(db)
			tt.mockBehavior(mock)

			err := cache.Set(context.Background(), tt.key, tt.value, tt.ttl)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		Foo string `json:"foo"`
	}

	tests := []struct {
		name          string
		key           string
		valueInRedis  interface{}
		expectedValue interface{}
		mockBehavior  func(mock redismock.ClientMock)
		expectedError error
	}{
		{
			name:         "Get successful",
			key:          "myKey",
			valueInRedis: testStruct{Foo: "bar"},
			expectedValue: testStruct{
				Foo: "bar",
			},
			mockBehavior: func(mock redismock.ClientMock) {
				val, _ := json.Marshal(testStruct{Foo: "bar"})
				mock.ExpectGetEx("myKey", 2*time.Hour).SetVal(string(val))
			},
			expectedError: nil,
		},
		{
			name:          "Get not found",
			key:           "missingKey",
			valueInRedis:  nil,
			expectedValue: testStruct{},
			mockBehavior: func(mock redismock.ClientMock) {
				mock.ExpectGetEx("missingKey", 2*time.Hour).RedisNil()
			},
			expectedError: redis.Nil,
		},
		{
			name:         "Get returns invalid JSON",
			key:          "invalidJSON",
			valueInRedis: "not-a-json",
			mockBehavior: func(mock redismock.ClientMock) {
				mock.ExpectGetEx("invalidJSON", 2*time.Hour).SetVal("not-a-json")
			},
			expectedError: errors.New("invalid character"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock := redismock.NewClientMock()
			cache := NewRedis(db)

			tt.mockBehavior(mock)

			var actual testStruct
			err := cache.Get(context.Background(), tt.key, &actual)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, actual)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisCache_Increment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		key           string
		incrResult    int64
		mockBehavior  func(mock redismock.ClientMock)
		expectedError error
	}{
		{
			name:       "Increment success",
			key:        "counter",
			incrResult: 42,
			mockBehavior: func(mock redismock.ClientMock) {
				mock.ExpectIncr("counter").SetVal(42)
			},
			expectedError: nil,
		},
		{
			name:       "Increment error",
			key:        "counter",
			incrResult: 0,
			mockBehavior: func(mock redismock.ClientMock) {
				mock.ExpectIncr("counter").SetErr(errors.New("redis failure"))
			},
			expectedError: errors.New("redis failure"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock := redismock.NewClientMock()
			cache := NewRedis(db)

			tt.mockBehavior(mock)

			val, err := cache.Increment(context.Background(), tt.key)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.incrResult, val)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
