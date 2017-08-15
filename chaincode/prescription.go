package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// using pointers is pointles.. cc won't give you back original element.
// have to physically add new element/struct instead of using pointers. Thanks fabric. >:(
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

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval))) //making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}

	var prescrip PrescriptionList
	jsonAsBytes, _ := json.Marshal(prescrip) //clear the PrescriptionList struct
	err = stub.PutState(PrescriptionsListStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Invoke - Our entry point to invoke a chaincode function (eg. write, createAccount, etc)
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
	} else if function == "add_prescription" { // called by doctor
		return t.add_prescription(stub, args)
	} else if function == "fill_prescription" { // called by pharmacist
		return t.fill_prescription(stub, args)
	}
	// else if function == "set_user" { //change owner of a marble
	// 	res, err := t.set_user(stub, args) //lets make sure all open trades are still valid
	// 	return res, err
	// }

	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query: " + function)
}

// ============================================================================================================================
// Read - read a variable from chaincode state (used by Query)
// ============================================================================================================================

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// ============================================================================================================================
// Write - Invoke function to write
// ============================================================================================================================
func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// add_prescription - creates a prescription and adds it to PrescriptionList struct. Called by doctors only
// ============================================================================================================================
func (t *SimpleChaincode) add_prescription(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0        1         2         3        4        5           6
	// "uid", "patient", "doctor", "drug", "dosage", "units", "description"
	if len(args) < 7 {
		return nil, errors.New("Incorrect number of arguments. Expecting 7")
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
		return nil, errors.New("5th argument (dosage) must be a numeric string")
	}

	/// what to do for boolean? numeric string?

	var prescription = Prescription{}
	prescription.Uid = args[0]
	prescription.Patient = args[1]
	prescription.Doctor = args[2]
	prescription.Drug = args[3]
	prescription.Dosage = dosage
	prescription.Units = args[5]
	prescription.Description = args[6]
	prescription.Filled = false // is this correct?
	prescription.Pharmacist = ""

	fmt.Println("below is new prescription: ")
	fmt.Println(prescription)

	////////////////////// 1) store prescription with Uid as key for easy search /////
	prescripAsBytes, _ := json.Marshal(prescription) //				             /////
	err = stub.PutState(args[0], prescripAsBytes)    // 						 /////
	if err != nil {                                  //                          /////
		return nil, err //														 /////
	} //																		 /////
	//////////////////////////////////////////////////////////////////////////////////

	//get the PrescriptionList struct
	PrescripListAsBytes, err := stub.GetState(PrescriptionsListStr)
	if err != nil {
		return nil, errors.New("Failed to get PrescriptionList")
	}
	var pl PrescriptionList
	json.Unmarshal(PrescripListAsBytes, &pl) //un stringify it aka JSON.parse()

	/////////////////////// 2) append prescription into marketplace //////////////////
	pl.List = append(pl.List, prescription)                    //		         /////
	fmt.Println("! appended prescription to PrescriptionList") //			     /////
	plAsBytes, _ := json.Marshal(pl)                           //				 /////
	err = stub.PutState(PrescriptionsListStr, plAsBytes)       //				 ///// rewrite marketplace
	if err != nil {                                            //                /////
		return nil, err //														 /////
	} //																		 /////
	//////////////////////////////////////////////////////////////////////////////////

	fmt.Println("- end of add_prescription")
	return nil, nil

}

// ============================================================================================================================
// fill_prescription - fills a prescription. Called by pharmacist only
// ============================================================================================================================
func (t *SimpleChaincode) fill_prescription(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0         1
	// uid",  "pharmacist"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	fmt.Println("- Beginning of fill_prescription")
	fmt.Println(args[0] + " - " + args[1])

	////////////////////// 1) update prescription itself /////////////////////////////
	// A: get prescription from blockchain				 						 /////
	prescripAsBytes, err := stub.GetState(args[0]) // 							 /////
	if err != nil {                                // 							 /////
		return nil, errors.New("Failed to get prescription") // 				 /////
	} // 																		 /////
	//																			 /////
	res := Prescription{}                 // 									 /////
	json.Unmarshal(prescripAsBytes, &res) //un stringify it aka JSON.parse()     /////
	//																			 /////
	// B: set prescription to filled and include pharmacist						 /////
	res.Filled = true                                //                          /////
	fmt.Println("! set filled prescription to true") // 						 /////
	fmt.Println(res)                                 // 						 /////
	//																			 /////
	res.Pharmacist = args[1]                     //        					     /////
	fmt.Println("! set prescription pharmacist") // 							 /////
	fmt.Println(res)                             // 							 /////
	//																			 /////
	// C: update prescription and push back into blockchain						 /////
	newPrescripAsBytes, _ := json.Marshal(res) // 								 /////
	//rewrite the prescription with id as key									 /////
	err = stub.PutState(args[0], newPrescripAsBytes) // 						 /////
	if err != nil {                                  // 						 /////
		return nil, err // 														 /////
	} // 																		 /////
	//////////////////////////////////////////////////////////////////////////////////

	//// 2. update prescription in PrescriptionList since we can't use pointers //////
	// A: get all active tasks in marketplace									 /////
	PrescripListAsBytes, err := stub.GetState(PrescriptionsListStr) //	 		 /////
	if err != nil {                                                 // 			 /////
		return nil, errors.New("Failed to get PrescriptionList") // 			 /////
	} // 																		 /////
	var pl PrescriptionList                  // 								 /////
	json.Unmarshal(PrescripListAsBytes, &pl) //un stringify it aka JSON.parse()  /////
	//																			 /////
	fmt.Print("Prescription List: ") // 										 /////
	fmt.Println(pl)                  // 										 /////
	//																			 /////
	// B: find prescription in list and update it								 /////
	for i := range pl.List { //iter through all the tasks						 /////
		fmt.Print("looking @ prescription: ") // 						 		 /////
		fmt.Println(pl.List[i])               // 								 /////
		//																		 /////
		if pl.List[i].Uid == args[0] { // found the trade to update 			 /////
			fmt.Println("Found prescription to fill") // 						 /////
			//																	 /////
			// t.modify_task(stub, []string{"add_submission", args[0], args[1]}) // add submission to single uid query
			//																	 /////
			pl.List[i].Filled = true                                        // 	 /////
			pl.List[i].Pharmacist = args[1]                                 // 	 /////
			fmt.Println("! filled prescription and added pharmacist pList") //	 /////
			fmt.Println(pl.List[i].Filled)                                  //	 /////
			fmt.Println(pl.List[i].Pharmacist)                              //	 /////
			//																	 /////
			// C: push filled prescription back into blockchain					 /////
			jsonAsBytes, _ := json.Marshal(pl)                     //			 /////
			err = stub.PutState(PrescriptionsListStr, jsonAsBytes) //rewrite the marketplace with new submission
			if err != nil {                                        //	 		 /////
				return nil, err //	 											 /////
			} //	 															 /////
			break //	 														 /////
		} else if i == (len(pl.List) - 1) { //	 								 /////
			return nil, errors.New("! Prescription not found in fill_prescription")
		} //	 																 /////
	} //	 																	 /////
	//////////////////////////////////////////////////////////////////////////////////

	fmt.Println("- end of fill_prescription")
	return nil, nil
}

// ============================================================================================================================
// Set Trade - create an open trade for a marble you want with marbles you have
// ============================================================================================================================
// func (t *SimpleChaincode) set_user(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
// 	var err error
// 	var toRes Account
// 	//     0         1        2        3         4         5
// 	// "fromUser", "500", "toUser", "reason", "hours", "comments"

// 	fmt.Println(args[0])
// 	fmt.Println(args[1])
// 	fmt.Println(args[2])
// 	fmt.Println(args[3])
// 	fmt.Println(args[4])
// 	fmt.Println(args[5])

// 	if len(args) < 5 {
// 		return nil, errors.New("Incorrect number of arguments. Expecting 6")
// 	}

// 	fromAccountAsBytes, err := stub.GetState(args[0])
// 	if err != nil {
// 		return nil, errors.New("Failed to get Sender")
// 	}
// 	toAccountAsBytes, err := stub.GetState(args[2])
// 	if err != nil {
// 		return nil, errors.New("Failed to get Receiver")
// 	}

// 	fromRes := Account{}
// 	json.Unmarshal(fromAccountAsBytes, &fromRes) //un stringify it aka JSON.parse()

// 	toRes = Account{}
// 	json.Unmarshal(toAccountAsBytes, &toRes)

// 	accountBalance := fromRes.GiveBalance

// 	transferAmount, err := strconv.Atoi(args[1])
// 	if err != nil {
// 		//Error because the amount entered is not a strNumber.
// 		// DO not need this case if we can get a number pad so user cannot enter other characters
// 		// handle error
// 		return nil, err
// 	}

// 	if accountBalance < transferAmount {
// 		fmt.Println("- Insufficient funds")
// 		return nil, errors.New("Failed to make Transaction - Insufficient funds")
// 	}

// 	toRes.PointsBalance = toRes.PointsBalance + transferAmount
// 	fromRes.GiveBalance = fromRes.GiveBalance - transferAmount

// 	toJsonAsBytes, _ := json.Marshal(toRes)
// 	err = stub.PutState(args[2], toJsonAsBytes) //rewrite the marble with id as key
// 	if err != nil {
// 		return nil, err
// 	}

// 	fromJsonAsBytes, _ := json.Marshal(fromRes)
// 	err = stub.PutState(args[0], fromJsonAsBytes) //rewrite the marble with id as key
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Println("Sucessful Transaction - end set trade")
// 	return nil, nil
// }
