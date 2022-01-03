package column_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahid-sohrabloo/chconn"
	"github.com/vahid-sohrabloo/chconn/column"
	"github.com/vahid-sohrabloo/chconn/setting"
)

func TestIPv6LC(t *testing.T) {
	t.Parallel()

	connString := os.Getenv("CHX_TEST_TCP_CONN_STRING")

	conn, err := chconn.Connect(context.Background(), connString)
	require.NoError(t, err)

	res, err := conn.Exec(context.Background(), `DROP TABLE IF EXISTS test_lc_ipv6`)
	require.NoError(t, err)
	require.Nil(t, res)
	settings := setting.NewSettings()
	settings.AllowSuspiciousLowCardinalityTypes(true)
	res, err = conn.ExecWithSetting(context.Background(), `CREATE TABLE test_lc_ipv6 (
				ipv6_lc LowCardinality(IPv6),
				ipv6_lc_nullable LowCardinality(Nullable(IPv6)),
				ipv6_lc_array Array(LowCardinality(IPv6)),
				ipv6_lc_array_nullable Array(LowCardinality(Nullable(IPv6)))
			) Engine=Memory`, settings)

	require.NoError(t, err)
	require.Nil(t, res)

	col := column.NewIPv6(false)
	colLC := column.NewLC(col)

	colNil := column.NewIPv6(true)
	colNilLC := column.NewLC(colNil)

	colArrayValues := column.NewIPv6(false)
	collArrayLC := column.NewLC(colArrayValues)
	colArray := column.NewArray(collArrayLC)

	colArrayValuesNil := column.NewIPv6(true)
	collArrayLCNil := column.NewLC(colArrayValuesNil)
	colArrayNil := column.NewArray(collArrayLCNil)

	var colInsert []net.IP
	var colInsertArray [][]net.IP
	var colInsertArrayNil [][]*net.IP
	var colNilInsert []*net.IP

	// var colMap

	rows := 10
	for i := 1; i <= rows; i++ {
		val := net.IP{
			byte(i), byte(i + 1), byte(i + 2), byte(i + 3),
			byte(i), byte(i + 1), byte(i + 2), byte(i + 3),
			byte(i), byte(i + 1), byte(i + 2), byte(i + 3),
			byte(i), byte(i + 1), byte(i + 2), byte(i + 3),
		}
		valArray := []net.IP{val, {
			byte(i), byte(i + 1), byte(i + 2), byte(i + 3),
			byte(i), byte(i + 1), byte(i + 2), byte(i + 3),
			byte(i), byte(i + 1), byte(i + 2), byte(i + 3),
			byte(i), byte(i + 1), byte(i + 2), byte(i + 3),
		}}
		valArrayNil := []*net.IP{&val, nil}

		col.AppendDict(val)
		colInsert = append(colInsert, val)

		// // example insert array
		colInsertArray = append(colInsertArray, valArray)
		colArray.AppendLen(len(valArray))
		for _, v := range valArray {
			colArrayValues.AppendDict(v)
		}

		// example insert nullable array
		colInsertArrayNil = append(colInsertArrayNil, valArrayNil)
		colArrayNil.AppendLen(len(valArrayNil))
		for _, v := range valArrayNil {
			colArrayValuesNil.AppendDictP(v)
		}

		// example add nullable
		if i%2 == 0 {
			colNilInsert = append(colNilInsert, &val)
			if i <= rows/2 {
				// example to add by poiner
				colNil.AppendDictP(&val)
			} else {
				// example to without poiner
				colNil.AppendDict(val)
			}
		} else {
			colNilInsert = append(colNilInsert, nil)
			if i <= rows/2 {
				// example to add by poiner
				colNil.AppendDictP(nil)
			} else {
				// example to add without poiner
				colNil.AppendDictNil()
			}
		}
	}

	insertstmt, err := conn.Insert(context.Background(), `INSERT INTO
		test_lc_ipv6(ipv6_lc,ipv6_lc_nullable,ipv6_lc_array,ipv6_lc_array_nullable)
	VALUES`)

	require.NoError(t, err)
	require.Nil(t, res)

	err = insertstmt.Commit(context.Background(),
		colLC,
		colNilLC,
		colArray,
		colArrayNil,
	)
	require.NoError(t, err)

	// example read all
	selectStmt, err := conn.Select(context.Background(), `SELECT ipv6_lc,
		ipv6_lc_nullable,ipv6_lc_array,ipv6_lc_array_nullable FROM
	test_lc_ipv6`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead := column.NewIPv6(false)
	colLCRead := column.NewLC(colRead)

	colNilRead := column.NewIPv6(true)
	colNilLCRead := column.NewLC(colNilRead)

	colArrayReadData := column.NewIPv6(false)
	colArrayLCRead := column.NewLC(colArrayReadData)
	colArrayRead := column.NewArray(colArrayLCRead)

	colArrayReadDataNil := column.NewIPv6(true)
	colArrayLCReadNil := column.NewLC(colArrayReadDataNil)
	colArrayReadNil := column.NewArray(colArrayLCReadNil)

	var colDataDict []net.IP
	var colDataKeys []int
	var colData []net.IP

	var colNilDataDict []net.IP
	var colNilDataKeys []int
	var colNilData []*net.IP

	var colArrayDataDict []net.IP
	var colArrayData [][]net.IP

	var colArrayDataDictNil []net.IP
	var colArrayDataNil [][]*net.IP

	var colArrayLens []int

	for selectStmt.Next() {
		err = selectStmt.NextColumn(colLCRead)
		require.NoError(t, err)
		colRead.ReadAll(&colDataDict)
		colLCRead.ReadAll(&colDataKeys)

		for _, k := range colDataKeys {
			colData = append(colData, colDataDict[k])
		}
		err = selectStmt.NextColumn(colNilLCRead)
		require.NoError(t, err)
		colNilRead.ReadAll(&colNilDataDict)
		colNilLCRead.ReadAll(&colNilDataKeys)

		for _, k := range colNilDataKeys {
			// 0 means nil
			if k == 0 {
				colNilData = append(colNilData, nil)
			} else {
				colNilData = append(colNilData, &colNilDataDict[k])
			}
		}

		// read array
		colArrayLens = colArrayLens[:0]
		err = selectStmt.NextColumn(colArrayRead)
		require.NoError(t, err)
		colArrayRead.ReadAll(&colArrayLens)
		colArrayReadData.ReadAll(&colArrayDataDict)
		for _, l := range colArrayLens {
			arr := make([]int, l)
			arrData := make([]net.IP, l)
			colArrayLCRead.Fill(arr)
			for i, k := range arr {
				arrData[i] = colArrayDataDict[k]
			}
			colArrayData = append(colArrayData, arrData)
		}

		// read array nil
		colArrayLens = colArrayLens[:0]
		err = selectStmt.NextColumn(colArrayReadNil)
		require.NoError(t, err)
		colArrayReadNil.ReadAll(&colArrayLens)
		colArrayReadDataNil.ReadAll(&colArrayDataDictNil)
		for _, l := range colArrayLens {
			arr := make([]int, l)
			arrData := make([]*net.IP, l)
			colArrayLCReadNil.Fill(arr)
			for i, k := range arr {
				// 0 means nil
				if k == 0 {
					arrData[i] = nil
				} else {
					arrData[i] = &colArrayDataDictNil[k]
				}
			}
			colArrayDataNil = append(colArrayDataNil, arrData)
		}
	}

	require.NoError(t, selectStmt.Err())

	assert.Equal(t, colInsert, colData)
	assert.Equal(t, colNilInsert, colNilData)
	assert.Equal(t, colInsertArray, colArrayData)
	assert.Equal(t, colInsertArrayNil, colArrayDataNil)

	selectStmt.Close()

	// example one by one
	selectStmt, err = conn.Select(context.Background(), `SELECT
		ipv6_lc,ipv6_lc_nullable FROM
	test_lc_ipv6`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead = column.NewIPv6(false)
	colLCRead = column.NewLC(colRead)

	colNilRead = column.NewIPv6(true)
	colNilLCRead = column.NewLC(colNilRead)

	colDataDict = colDataDict[:0]
	colData = colData[:0]

	colNilDataDict = colNilDataDict[:0]
	colNilData = colNilData[:0]

	for selectStmt.Next() {
		err = selectStmt.NextColumn(colLCRead)
		require.NoError(t, err)
		colRead.ReadAll(&colDataDict)

		for colLCRead.Next() {
			colData = append(colData, colDataDict[colLCRead.Value()])
		}
		err = selectStmt.NextColumn(colNilLCRead)
		require.NoError(t, err)
		colNilRead.ReadAll(&colNilDataDict)

		for colNilLCRead.Next() {
			k := colNilLCRead.Value()
			// 0 means nil
			if k == 0 {
				colNilData = append(colNilData, nil)
			} else {
				colNilData = append(colNilData, &colNilDataDict[k])
			}
		}
	}

	require.NoError(t, selectStmt.Err())

	selectStmt.Close()

	assert.Equal(t, colInsert, colData)
	assert.Equal(t, colNilInsert, colNilData)
	conn.Close(context.Background())
}
