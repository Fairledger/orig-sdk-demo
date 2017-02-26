/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var contractIndexStr = "_contractindex"		// name for key/value that will store the list of contracts

type ContractTerms struct{
	Max_Temperature_F int `json:"max_temperature_f"`
	Product_Type string `json:"product"`

}

type Shipment struct{
	ContractID string `json:"contractID"`
	Value	int `json:"value_dollars"`
	Temperature_F int `json:"temerature_f"`
	CarrierName string `json:"carrier_name"`
	Location string `json:"location"`
	ShipEvent string `json:"shipEvent"`
	Timestamp int64 `json:"timestamp"`
}

var EVENT_COUNTER = "event_counter"

// ================================================================================
// Run - Our entry point for Invocations 
// ================================================================================
/*func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}*/

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int // Asset holdings
	var err error

	fmt.Println("HERERE I'm in Init()")
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(contractIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

  err = stub.PutState(EVENT_COUNTER, []byte("1"))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, funct string, args []string) ([]byte, error) {
	fmt.Println("HERERERER")
	fmt.Println("invoke is running " + funct)

	// Handle different functions
	if funct == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if funct == "init_contract_terms" {				//create a business contract 
		return t.init_terms(stub, args)
	} 
	/*else if function == "shipment_event" {				//Enter the shipment event within the supply chain route 
		return t.shipment_event(stub, args)
	} else if function == "transfer_funds" {		  //transfer funds from one participant to another
		res, err := t.transfer_funds(stub, args)
		//cleanTrades(stub)													//lets make sure all open trades are still valid
		return res, err
	}*/
	fmt.Println("invoke did not find func: " + funct)					//error

	//Event based
  b, err := stub.GetState(EVENT_COUNTER)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	noevts, _ := strconv.Atoi(string(b))

	tosend := "Invoke() Event Eounter is " + string(b)

	err = stub.PutState(EVENT_COUNTER, []byte(strconv.Itoa(noevts+1)))
	if err != nil {
		return nil, err
	}

	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return nil, err
  }

	return nil, errors.New("Received unknown function invocation")


}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) init_terms(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

	var err error

	// 0								1
	// "product_type"	"max_temperature_f"
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	//input sanitation
	fmt.Println("- start init contract terms")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}

	product_type			:= args[0]			// type of product being transferred  
	max_temperature_f, err := strconv.Atoi(args[1])	// max temperature
	if err != nil {
		return nil, errors.New("2nd argument must be a numeric string")
	}

	// Get the state from the ledger
	contractAsBytes, err := stub.GetState(product_type)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}

	res := ContractTerms{}
	json.Unmarshal(contractAsBytes, &res)
	if res.Product_Type == product_type {
		retstr := "Terms of Contract for product " + res.Product_Type + " already exists"
		return nil, errors.New(retstr)
	}

	//build the marble json string manually
	str := `{"product_type": "` + product_type + `", "max_temperature_f": ` + strconv.Itoa(max_temperature_f) + `"}`
	err = stub.PutState(product_type, []byte(str))						//store contracty with product type as key
	if err != nil {
		return nil, err
	}
		//get the marble index
	contractsAsBytes, err := stub.GetState(contractIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get contract terms index")
	}
	var contractIndex []string
	json.Unmarshal(contractsAsBytes, &contractIndex)							//un stringify it aka JSON.parse()
	
	//append
	contractIndex = append(contractIndex, product_type)									//add marble name to index list
	fmt.Println("! contract index: ", contractIndex)
	jsonAsBytes, _ := json.Marshal(contractIndex)
	err = stub.PutState(contractIndexStr, jsonAsBytes)						//store name of marble

	fmt.Println("- end init contract terms")

	//Event based
        b, err := stub.GetState(EVENT_COUNTER)
	if err != nil {
		return nil, errors.New("Failed to get state")
	}
	noevts, _ := strconv.Atoi(string(b))

	tosend := "init_terms Event Counter is " + string(b)

	err = stub.PutState(EVENT_COUNTER, []byte(strconv.Itoa(noevts+1)))
	if err != nil {
		return nil, err
	}

	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return nil, err
        }
	return nil, nil
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	return nil, nil
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var A string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return nil, errors.New(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return Avalbytes, nil
}

// ================================================================================
// Main
// ================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Supply Chain chaincode: %s", err)
	}
}

