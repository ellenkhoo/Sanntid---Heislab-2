package elevator

import "fmt"

//definierer enum for Button og Dirn 
type Button int 
type Dirn int 

const (
	D_Up Dirn = iota
	D_Down 
	D_Stop
)

//funksjonen som stimulerer initialisering av heismaskinvare
func init(){
	elevatorHardwareInit()
}

//simulerer initialisering av heisens maskinvare
func elevatorHardwareInit(){
	fmt.Println("Initialising elevator hardware...")
	//her kan du vi legge til kode for maskinvareinitialisering
}

//funksjonen som simulerer å hente signalet fra en knapp på en spesifikk etasje
func wrapRequestButton(f int, b Button) int{
	return elevatorHardwareGetButtonSignal(b, f)
}

//simulerer å hente knappesignal for en spesifikk etasje og knapp
func elevatorHardwareGetButtonSignal(b Button, f int) int{
	fmt.Printf("Getting button signal for floor %d and button %d\n", f, b)
	return 1 //simulerer at signalet er aktivert
}

//funksjonen for å sette en lampe på en knapp (tillater å sette den på eller av)
func wrapRequestButtonLight(f int, b Button, v int){
	elevatorHardwareSetButtonLamp(b, f, v)
}

//simulerer å sette lampen til en spesifikk verdi for en knapp på en etasje
func elevatorHardwareSetButtonLamp(b Button, f int, v int){
	fmt.Printf("Setting button light for florr %d, buttin %d, value %d\n", f, b, v)
	//her kan vi legge til kode som setter lampeverdien på en knapp
}

//funksjonen som simulerer å sette motorretning
func wrapMotorDirection(d Drin){
	elevatorHardwareSetMotorDirection(d)
}

//simulerer å sette motorretningen til heisen
func elevatorHardwareSetMotorDirection(d Dirn){
	fmt.Printf("Setting motor direction to %d\n", d)
	//her kan vi legge til kode som setter motorens retning
}