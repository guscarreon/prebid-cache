package backends

import (
	"context"
	"errors"
	"testing"

	"github.com/gocql/gocql"
	"github.com/prebid/prebid-cache/utils"

	"github.com/stretchr/testify/assert"
)

func TestCassandraClientGet(t *testing.T) {
	cassandraBackend := &CassandraBackend{}

	type testInput struct {
		cassandraClient CassandraDB
		key             string
	}

	type testExpectedValues struct {
		value string
		err   error
	}

	testCases := []struct {
		desc     string
		in       testInput
		expected testExpectedValues
	}{
		{
			"CassandraBackend.Get() throws a Cassandra ErrNotFound error",
			testInput{
				NewErrorProneCassandraClient(gocql.ErrNotFound),
				"someKeyThatWontBeFound",
			},
			testExpectedValues{
				value: "",
				err:   utils.KeyNotFoundError{},
			},
		},
		{
			"CassandraBackend.Get() throws an error different from Cassandra ErrNotFound error",
			testInput{
				NewErrorProneCassandraClient(errors.New("some other get error")),
				"someKey",
			},
			testExpectedValues{
				value: "",
				err:   errors.New("some other get error"),
			},
		},
		{
			"CassandraBackend.Get() doesn't throw an error",
			testInput{
				NewGoodCassandraClient("defaultKey", "aValue"),
				"defaultKey",
			},
			testExpectedValues{
				value: "aValue",
				err:   nil,
			},
		},
	}

	for _, tt := range testCases {
		cassandraBackend.client = tt.in.cassandraClient

		// Run test
		actualValue, actualErr := cassandraBackend.Get(context.TODO(), tt.in.key)

		// Assertions
		assert.Equal(t, tt.expected.value, actualValue, tt.desc)
		assert.Equal(t, tt.expected.err, actualErr, tt.desc)
	}
}

func TestCassandraClientPut(t *testing.T) {
	cassandraBackend := &CassandraBackend{
		defaultTTL: 50,
	}

	type testInput struct {
		cassandraClient CassandraDB
		key             string
		valueToStore    string
		ttl             int
	}

	type testExpectedValues struct {
		value string
		err   error
	}

	testCases := []struct {
		desc     string
		in       testInput
		expected testExpectedValues
	}{
		{
			"CassandraBackend.Put() throws error",
			testInput{
				NewErrorProneCassandraClient(gocql.ErrNoConnections),
				"someKey",
				"someValue",
				10,
			},
			testExpectedValues{
				"",
				errors.New("gocql: no hosts available in the pool"),
			},
		},
		{
			"CassandraBackend.Put() gets called with zero ttlSeconds, value gets successfully set anyways",
			testInput{
				NewGoodCassandraClient("defaultKey", "aValue"),
				"defaultKey",
				"aValue",
				0,
			},
			testExpectedValues{
				"aValue",
				nil,
			},
		},
		{
			"CassandraBackend.Put() successful, no need to set defaultTTL because ttl is greater than zero",
			testInput{
				NewGoodCassandraClient("defaultKey", "aValue"),
				"defaultKey",
				"aValue",
				1,
			},
			testExpectedValues{
				"aValue",
				nil,
			},
		},
	}

	for _, tt := range testCases {
		// Assign aerospike backend cient
		cassandraBackend.client = tt.in.cassandraClient

		// Run test
		actualErr := cassandraBackend.Put(context.TODO(), tt.in.key, tt.in.valueToStore, tt.in.ttl)

		// Assert Put error
		assert.Equal(t, tt.expected.err, actualErr, tt.desc)

		// Assert value
		if tt.expected.err == nil {
			storedValue, getErr := cassandraBackend.Get(context.TODO(), tt.in.key)

			assert.NoError(t, getErr, tt.desc)
			assert.Equal(t, tt.expected.value, storedValue, tt.desc)
		}
	}
}