/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

var PrescriptionsListStr = "_prescriptions" // name for the key/value that will store all prescriptions

type Prescription struct {
	Uid         string `json:"id"`
	Patient     string `json:"patient"` // patients are defined by their full name
	Doctor      string `json:"doctor"`  // doctor who prescribed the prescription
	Drug        string `json:"drug"`    // name of the drug prescribed
	Dosage      int    `json:"dosage"`  // ex. 450
	Units       string `json:"units"`   // ex. ml
	Description string `json:"description"`
	Filled      bool   `json:"filled"`     // 0 = not filled, 1 = filled
	Pharmacist  string `json:"pharmacist"` // pharmacist who filled the prescription
}

type PrescriptionList struct {
	List []Prescription `json:"prescriptionList"`
}

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	var Aval int
	var err error

	_, args := stub.GetFunctionAndParameters()

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval))) //making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return shim.Error("Init errror")
	}

	var prescrip PrescriptionList
	jsonAsBytes, _ := json.Marshal(prescrip) //clear the PrescriptionList struct
	err = stub.PutState(PrescriptionsListStr, jsonAsBytes)
	if err != nil {
		return shim.Error("Init errror")
	}
	return shim.Success(nil)
}

// ============================================================================================================================
// Invoke - Our entry point to invoke a chaincode function (eg. write, createAccount, etc)
// ============================================================================================================================
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "write" {
		return t.write(stub, args)
	} else if function == "add_prescription" { // called by doctor
		return t.add_prescription(stub, args)
	} else if function == "fill_prescription" { // called by pharmacist
		return t.fill_prescription(stub, args)
	} else if function == "query" {
		// queries an entity state
		return t.query(stub, args)
	}

	fmt.Println("invoke did not find func: " + function)
	return shim.Error("Received unknown function invocation: ")

}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleAsset) query(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("query is running ")

	_, args = stub.GetFunctionAndParameters()

	var key string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("json error")
	}

	return shim.Success(valAsbytes)
}

// ============================================================================================================================
// Write - Invoke function to write
// ============================================================================================================================
func (t *SimpleAsset) write(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2: name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return shim.Error("jSome errror")
	}
	return shim.Success(nil)

}

// ============================================================================================================================
// add_prescription - creates a prescription and adds it to PrescriptionList struct. Called by doctors only
// ============================================================================================================================
func (t *SimpleAsset) add_prescription(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	//   0        1         2         3        4        5           6
	// "uid", "patient", "doctor", "drug", "dosage", "units", "description"
	if len(args) < 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	fmt.Println("- Beginning of add_prescription")

	fmt.Println("Number of args: ")
	fmt.Println(len(args))
	fmt.Println(args[0])
	fmt.Println(args[1])
	fmt.Println(args[2])
	fmt.Println(args[3])
	fmt.Println(args[4])
	fmt.Println(args[5])
	fmt.Println(args[6])

	dosage, err := strconv.Atoi(args[4])
	if err != nil {
		return shim.Error("5th argument (dosage) must be a numeric string")
	}

	var prescription = Prescription{}
	prescription.Uid = args[0]
	prescription.Patient = args[1]
	prescription.Doctor = args[2]
	prescription.Drug = args[3]
	prescription.Dosage = dosage
	prescription.Units = args[5]
	prescription.Description = args[6]
	prescription.Filled = false
	prescription.Pharmacist = ""

	fmt.Println("below is new prescription: ")
	fmt.Println(prescription)

	/////////// 1) store prescription with Uid as key for easy search
	prescripAsBytes, _ := json.Marshal(prescription)
	err = stub.PutState(args[0], prescripAsBytes)
	if err != nil {
		return shim.Error("Error pushing prescription back into blockchain")
	}

	/////////// 2) append prescription into prescriptionList
	// A: get the PrescriptionList struct
	PrescripListAsBytes, err := stub.GetState(PrescriptionsListStr)
	if err != nil {
		return shim.Error("Failed to get PrescriptionList")
	}
	var pl PrescriptionList
	json.Unmarshal(PrescripListAsBytes, &pl) //un stringify it aka JSON.parse()

	// B: append prescription into prescription List
	pl.List = append(pl.List, prescription)
	fmt.Println("! appended prescription to PrescriptionList")

	// C: push prescription list back into blockchain
	plAsBytes, _ := json.Marshal(pl)
	err = stub.PutState(PrescriptionsListStr, plAsBytes) // rewrite prescription list
	if err != nil {
		return shim.Error("Error pushing prescription list back into blockchain")
	}

	fmt.Println("- end of add_prescription")
	return shim.Success(nil)
}

// ============================================================================================================================
// fill_prescription - fills a prescription. Called by pharmacist only
// ============================================================================================================================
func (t *SimpleAsset) fill_prescription(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	//   0         1
	// uid",  "pharmacist"
	if len(args) < 2 {
		return shim.Error("incorrect number of arguments. Expecting 2.")
	}

	fmt.Println("- Beginning of fill_prescription")
	fmt.Println(args[0] + " - " + args[1])

	///////// 1) update prescription itself
	// A: get prescription from blockchain
	prescripAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Error in getting prescription from blockchain")
	}
	if len(prescripAsBytes) == 0 {
		return shim.Error("Prescription not found. This is not a legitimate prescription.")
	}

	res := Prescription{}
	json.Unmarshal(prescripAsBytes, &res) //un stringify it aka JSON.parse()

	if res.Filled == false {

		// B: set prescription to filled and include pharmacist
		res.Filled = true
		fmt.Println("! set filled prescription to true")
		fmt.Println(res)

		res.Pharmacist = args[1]
		fmt.Println("! set prescription pharmacist")
		fmt.Println(res)

		// C: update prescription and push back into blockchain
		newPrescripAsBytes, _ := json.Marshal(res)
		//rewrite the prescription with id as key
		err = stub.PutState(args[0], newPrescripAsBytes)
		if err != nil {
			return shim.Error("Error in rewriting prescription into blockchain.")
		}

		//// 2. update prescription in PrescriptionList since we can't use pointers
		// A: get all active tasks in marketplace
		PrescripListAsBytes, err := stub.GetState(PrescriptionsListStr)
		if err != nil {
			return shim.Error("Failed to get PrescriptionList")
		}
		var pl PrescriptionList
		json.Unmarshal(PrescripListAsBytes, &pl) //un stringify it aka JSON.parse()

		fmt.Print("Prescription List: ")
		fmt.Println(pl)

		// B: find prescription in list and update it
		for i := range pl.List { //iter through all the tasks
			fmt.Print("looking @ prescription: ")
			fmt.Println(pl.List[i])

			if pl.List[i].Uid == args[0] { // found the trade to update
				fmt.Println("Found prescription to fill")

				pl.List[i].Filled = true
				pl.List[i].Pharmacist = args[1]
				fmt.Println("! filled prescription and added pharmacist pList")
				fmt.Println(pl.List[i].Filled)
				fmt.Println(pl.List[i].Pharmacist)

				// C: push filled prescription back into blockchain
				jsonAsBytes, _ := json.Marshal(pl)
				err = stub.PutState(PrescriptionsListStr, jsonAsBytes) // rewrite the presciption list with new prescription
				if err != nil {
					return shim.Error("Error in rewriting prescription to list")
				}
				break
			} else if i == (len(pl.List) - 1) {
				return shim.Error("! Prescription not found in fill_prescription")
			}
		}

		fmt.Println("- end of fill_prescription")

		return shim.Success(nil)

	}

	fmt.Println("- end of fill_prescription")

	return shim.Error("Prescription was previously filled. This is a double fill attempt.")

}

// ============================================================================================================================
// Main - main function
// ============================================================================================================================
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
