package main

// Imports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim" // ①
	// peerパッケージのインポート時にscをエイリアスとして定義している。これによってfunctionのレスポンスをsc.Responseと書くことができる。
  // 以下のソースコードのようにエイリアスとして定義しなければ、functionのreturnはpeer.Responseと定義する。
  // https://github.com/bcken001a/fabric-samples/blob/release/chaincode/sacc/sacc.go
	sc "github.com/hyperledger/fabric/protos/peer" // ②
)

type SmartContract struct {
}

// Car Structの定義。jsonで出力したいので、`json:"○○"`としている。
type Car struct {
	Make   string `json:"make"`
	Model  string `json:"model"`
	Colour string `json:"colour"`
	Owner  string `json:"owner"`
}

// 初期化
// 何も処理を行わないので、shim.Successをそのままreturnしている。
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// チェーンコードのInvoke
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// 要求されたスマートコントラクトの関数と引数を取得する
	// function: コマンドで要求したコントラクトの関数
	// args: コントラクト実行時に必要な変数
	function, args := APIstub.GetFunctionAndParameters()

	if function == "queryCar" {
		return s.queryCar(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createCar" {
		return s.createCar(APIstub, args)
	} else if function == "queryAllCars" {
		return s.queryAllCars(APIstub)
	} else if function == "changeCarOwner" {
		return s.changeCarOwner(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryCar(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// その関数にどんな引数が必要なのかはっきりさせるために、チェーンコード内に以下のようにコメントを残しておくと良い
	//  0
	// "id"

	// エラーチェック
	// 変数の数が0でなかったらエラー
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// 変数をcarAsBytesに代入
	// GetStateメソッドはgithubからimportしたfabricのパッケージ①の中で定義されているメソッド
	// https://github.com/bcken001a/fabric/blob/release/core/chaincode/shim/mockstub.go
	carAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(carAsBytes)
}

//レッジャーの初期化を行う関数
// 上で定義したtype Car structと一致するように定義する。
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	cars := []Car{
		Car{Make: "Toyota", Model: "Prius", Colour: "blue", Owner: "Tomoko"},
		Car{Make: "Ford", Model: "Mustang", Colour: "red", Owner: "Brad"},
		Car{Make: "Hyundai", Model: "Tucson", Colour: "green", Owner: "Jin Soo"},
		Car{Make: "Volkswagen", Model: "Passat", Colour: "yellow", Owner: "Max"},
		Car{Make: "Tesla", Model: "S", Colour: "black", Owner: "Adriana"},
		Car{Make: "Peugeot", Model: "205", Colour: "purple", Owner: "Michel"},
		Car{Make: "Chery", Model: "S22L", Colour: "white", Owner: "Aarav"},
		Car{Make: "Fiat", Model: "Punto", Colour: "violet", Owner: "Pari"},
		Car{Make: "Tata", Model: "Nano", Colour: "indigo", Owner: "Valeria"},
		Car{Make: "Holden", Model: "Barina", Colour: "brown", Owner: "Shotaro"},
	}

	i := 0
	for i < len(cars) {
		fmt.Println("i is ", i)
		// carsの情報をMarshalでエンコードする。
		carAsBytes, _ := json.Marshal(cars[i])
		// PutStateもgithubからimportしたfabricのパッケージ①の中で定義されているメソッド
		// 第一引数がKeyで、第二引数がValueとして登録される
		// この場合はCAR1、CAR2、CAR3...がKeyとなる
		APIstub.PutState("CAR"+strconv.Itoa(i), carAsBytes)
		fmt.Println("Added", cars[i])
		i = i + 1
	}

	return shim.Success(nil)
}

// 新しくCarを作成する関数
func (s *SmartContract) createCar(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//    0      1        2       3
	// "make" "model" "colour" "owner"

	// エラーチェック
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	// コマンドの引数を使って、作成するCar型の配列を作る
	var car = Car{Make: args[1], Model: args[2], Colour: args[3], Owner: args[4]}

	// エンコード
	carAsBytes, _ := json.Marshal(car)
	// レッジャーに登録する
	APIstub.PutState(args[0], carAsBytes)

	return shim.Success(nil)
}

// 全てのCar情報を紹介する関数
func (s *SmartContract) queryAllCars(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "CAR0"
	endKey := "CAR999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllCars:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// Carのownerを変更する関数
func (s *SmartContract) changeCarOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	//   0      1
	// "key" "owner"

	// エラーチェック
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	carAsBytes, _ := APIstub.GetState(args[0])
	car := Car{}

	// GetStateで取得した情報（carAsBytes）はbytes形式なので、Unmarshalでデコードする。
	json.Unmarshal(carAsBytes, &car)
	// 取得したcarのownerをコマンドの引数で定義したownerに変更する
	car.Owner = args[1]

	// 再び、エンコードする
	carAsBytes, _ = json.Marshal(car)
	// レッジャーに登録する
	APIstub.PutState(args[0], carAsBytes)

	return shim.Success(nil)
}

func main() {

	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
