package main

import (
	"contactApp/user"
	"fmt"
)

func main() {

	// Create Admin user
	admin := user.NewAdmin("Alice", "Johnson")
	fmt.Println("Admin created:")
	admin.PrintDetails()

	// Create Staff user using Admin
	staff1,errr := admin.NewStaff("Bob", "Smith")
	if errr==nil{
		fmt.Println("Staff created:")
		staff1.PrintDetails()

	}else{
		fmt.Println(errr)
	}

	staff2,errr := admin.NewStaff("Charlie", "Brown")
	if errr==nil{
		fmt.Println("Staff created:")
		staff2.PrintDetails()

	}else{
		fmt.Println(errr)
	}
	

	// Admin reads all users
	fmt.Println("\nAdmin reading all users:")
	users, err := admin.ReadUsers()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		for _, u := range users {
			u.PrintDetails()
		}
	}

	// Update User by parameter (Admin updates first name of a staff)
	fmt.Println("\nAdmin updating Staff user (changing Bob's firstname to 'Robert'):")
	err = staff1.UpdateUserByParameter("firstname", "Robert")
	if err != nil {
		fmt.Println("Error updating user:", err)
	} else {
		staff1.PrintDetails()
	}

	// Admin reads all users
	fmt.Println("\nAdmin reading all users:")
	users, err = admin.ReadUsers()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		for _, u := range users {
			u.PrintDetails()
		}
	}

	// Admin deletes a staff user
	fmt.Println("\nAdmin deleting a Staff user (Charlie):")
	err = admin.DeleteUser(staff2.UserID)
	if err != nil {
		fmt.Println("Error deleting user:", err)
	} else {
		fmt.Println("Staff user deleted.")
	}

	// Admin reads all users
	fmt.Println("\nAdmin reading all users:")
	users, err = admin.ReadUsers()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		for _, u := range users {
			u.PrintDetails()
		}
	}

	// Check if the deleted user is inactive
	fmt.Println("\nChecking if Charlie is deactivated:")
	staff2.PrintDetails()

	// Admin tries to delete an inactive user (Charlie)
	fmt.Println("\nAdmin trying to delete inactive user (Charlie) again:")
	err = admin.DeleteUser(staff2.UserID)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Invalid Update User by parameter
	fmt.Println("\nTrying to update user with invalid parameter:")
	err = staff1.UpdateUserByParameter("invalidParam", "SomeValue")
	if err != nil {
		fmt.Println("Error updating user:", err)
	}
	

	// Staff creates a new contact
	err = staff1.CreateContact("John", "Doe")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Contact created by Staff:")
		staff1.PrintDetails()
	}

	// Update contact
	err = staff1.UpdateContact(0, "John", "Doe Jr.")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Contact updated by Staff:")
		staff1.PrintDetails()
	}


	//Delete contact
	err = staff1.DeleteContact(0)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Contact deleted by Staff:")
		staff1.PrintDetails()
	}


	
	// // Staff creates a new contact detail
	// err = staff1.CreateContactDetail(0, "Phone", "123-456-7890")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// } else {
	// 	fmt.Println("Contact Detail added:")
	// 	staff1.PrintDetails()
	// }

}
