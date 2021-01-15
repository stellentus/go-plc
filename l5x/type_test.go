package l5x

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStructType(name string, nts []NamedType) structType {
	str, err := newStructType(name, nts)
	if err != nil {
		panic("Test contains invalid name '" + name + "'")
	}
	return str
}

func newTestMember(varName, ty string) Member {
	return Member{
		Name:     varName,
		DataType: ty,
	}
}

func ExampleNamedTypeDeclaration() {
	fmt.Println(NamedTypeDeclaration(NamedType{
		GoName: "MY_VAR",
		Type:   typeBOOL,
	}))
	fmt.Println(NamedTypeDeclaration(NamedType{
		GoName: "MY_STRUCT",
		Type:   newTestStructType("StructType", nil), // The nil members don't matter
	}))
	// Output:
	// MY_VAR bool
	// MY_STRUCT StructType
}

func ExampleTypeDefinition() {
	types := []Type{}
	types = append(types, typeBOOL) // prints nothing as it's built-in
	types = append(types, newTestStructType("DemoStruct", []NamedType{
		{GoName: "VAR", Type: typeBOOL},
	}))
	types = append(types, newTestStructType("FancyThing", []NamedType{
		{GoName: "COUNT", Type: typeINT},
		{GoName: "dsInstance", Type: types[1]},
	}))

	for _, ty := range types {
		fmt.Println(TypeDefinition(ty))
	}

	// Output:
	//
	// type DemoStruct struct {
	// 	VAR bool
	// }
	// type FancyThing struct {
	// 	COUNT int16
	// 	dsInstance DemoStruct
	// }
}

func ExampleTypeList_WriteDefinitions() {
	err := expectedTypeList.WriteDefinitions(os.Stdout)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	// Output:
	// type TIMER struct {
	// 	PRE int32
	// 	ACC int32
	// 	EN bool
	// 	TT bool
	// 	DN bool
	// }
	//
	// type COUNTER struct {
	// 	PRE int32
	// 	ACC int32
	// 	CU bool
	// 	CD bool
	// 	DN bool
	// 	OV bool
	// 	UN bool
	// }
	//
	// type Dow struct {
	// 	DayOW int16
	// 	Month int32
	// 	MonthCode [13]int32
	// 	DayOW1 float32
	// }
	//
	// type Big_data_type struct {
	// 	XprivateX_cleaning_c0 int8
	// 	CLEAN_MODE int16
	// 	XprivateX_cleaning_c7 int8
	// }
	//
	// type Datas_for_eating struct {
	// 	TIMER
	// 	XprivateX_cleaning_c0 int8
	// 	FOOD_TIMER TIMER
	// 	MEAL_PREP_TIMER TIMER
	// 	BHAIG29GI TIMER
	// 	COUNTDOWN_TO_DESSERT TIMER
	// 	STEPS_REQUIRED int16
	// 	SoMuchData Big_data_type `plc:"soMuchData"`
	// }
}

func TestMemberAsNamedTypeBasic(t *testing.T) {
	type memberAsNamedTypeTestData struct {
		PlcName      string
		Member              // input
		Name, GoName string // expected
	}

	newMemberAsNamedTypeTestDataBasic := func(plcTy, goTy string) memberAsNamedTypeTestData {
		varName := "Test_varNAME" + plcTy + goTy // Arbitrary name
		return memberAsNamedTypeTestData{
			PlcName: plcTy,
			Member:  newTestMember(varName, plcTy),
			Name:    varName,
			GoName:  goTy,
		}
	}

	tests := []memberAsNamedTypeTestData{
		newMemberAsNamedTypeTestDataBasic("BOOL", "bool"),
		newMemberAsNamedTypeTestDataBasic("SINT", "int8"),
		newMemberAsNamedTypeTestDataBasic("INT", "int16"),
		newMemberAsNamedTypeTestDataBasic("DINT", "int32"),
		newMemberAsNamedTypeTestDataBasic("LINT", "int64"),
		newMemberAsNamedTypeTestDataBasic("USINT", "uint8"),
		newMemberAsNamedTypeTestDataBasic("UINT", "uint16"),
		newMemberAsNamedTypeTestDataBasic("UDINT", "uint32"),
		newMemberAsNamedTypeTestDataBasic("ULINT", "uint64"),
		newMemberAsNamedTypeTestDataBasic("REAL", "float32"),
		newMemberAsNamedTypeTestDataBasic("LREAL", "float64"),
		newMemberAsNamedTypeTestDataBasic("STRING", "string"),
		newMemberAsNamedTypeTestDataBasic("BYTE", "byte"),
		newMemberAsNamedTypeTestDataBasic("WORD", "uint16"),
		newMemberAsNamedTypeTestDataBasic("DWORD", "uint32"),
		newMemberAsNamedTypeTestDataBasic("LWORD", "uint64"),
	}

	for _, test := range tests {
		t.Run(test.PlcName, func(t *testing.T) {
			nt, err := test.Member.AsNamedType(NewTypeList())
			assert.NoError(t, err)
			assert.Equal(t, test.Name, nt.GoName)
			assert.Equal(t, test.PlcName, nt.Type.PlcName())
			assert.Equal(t, test.GoName, nt.Type.GoName())
			assert.Equal(t, "", nt.Type.GoTypeString()) // These basic type don't have a definition
		})
	}
}

func TestMemberAsNamedTypeError(t *testing.T) {
	tests := []struct {
		TestName string
		Member
	}{
		{"UnknownType", newTestMember("MyVar", "UNKNOWN_TYPE")},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			_, err := test.Member.AsNamedType(NewTypeList())
			assert.Error(t, err)
		})
	}
}

func TestTypeListWithPlcName(t *testing.T) {
	tests := []struct {
		PlcName string
		Type
	}{
		{"LINT", typeLINT},
		{"USINT", typeUSINT},
		{"UINT", typeUINT},
	}

	for _, test := range tests {
		t.Run(test.PlcName, func(t *testing.T) {
			typ, err := NewTypeList().WithPlcName(test.PlcName)
			require.NoError(t, err)
			assert.Equal(t, test.Type.PlcName(), typ.PlcName())
			assert.Equal(t, test.Type.GoName(), typ.GoName())
			assert.Equal(t, test.Type.GoTypeString(), typ.GoTypeString())
		})
	}
}

func newTestDataType() DataType {
	return DataType{
		Name:    "ExampleDataType",
		Members: []Member{newTestMember("VarName", "INT")},
	}
}

func dtToStringDT(dt *DataType) {
	dt.Family = DataTypeFamilyString
	dt.Members = []Member{
		newTestMember("LEN", "DINT"),
		newTestMember("DATA", "SINT"),
	}
	dt.Members[1].Dimension = 15
	dt.Members[1].Radix = RadixASCII
}

func TestDataTypeAsNamedType(t *testing.T) {
	runTest := func(name, goString string, dtMod func(dt *DataType)) {
		t.Run(name, func(t *testing.T) {
			dt := newTestDataType()
			dt.Name = name
			dtMod(&dt) // Allows caller to modify dt make it more complicated
			nt, err := dt.AsType(NewTypeList())
			assert.NoError(t, err)
			assert.Equal(t, name, nt.PlcName(), "PlcName should match")
			assert.Equal(t, name, nt.GoName(), "GoName should match")
			assert.Equal(t, goString, nt.GoTypeString(), "GoTypeString should match")
		})
	}

	runTest(
		"BasicTest",
		"struct {\n\tVarName int16\n}",
		func(dt *DataType) {},
	)
	runTest(
		"SimpleStruct",
		"struct {\n\tVarName int16\n\tMY_VAR float32\n\tOtherVar int8 `plc:\"otherVar\"`\n}",
		func(dt *DataType) {
			dt.Members = append(dt.Members, newTestMember("MY_VAR", "REAL"))
			dt.Members = append(dt.Members, newTestMember("otherVar", "SINT"))
		},
	)
	runTest(
		"SkipBit",
		"struct {\n\tVarName int16\n}",
		func(dt *DataType) {
			dt.Members = append(dt.Members, newTestMember("otherVar", "BIT"))
		},
	)
	runTest(
		"String",
		"[15]int8",
		dtToStringDT,
	)
	runTest(
		"SimpleStruct",
		"struct {\n\tVarName int16\n\tBad_name__ float32 `plc:\"bad name 💔\"`\n}",
		func(dt *DataType) {
			dt.Members = append(dt.Members, newTestMember("bad name 💔", "REAL"))
		},
	)
}

func TestDataTypeAsNamedTypePredeclared(t *testing.T) {
	knownTypes := NewTypeList()
	err := knownTypes.AddControlLogixTypes()
	require.NoError(t, err, "There should be no issue adding ControlLogix types to a NewTypeList")

	embType := newTestDataType()
	embType.Name = "RegisteredType"
	embNT, err := embType.AsType(knownTypes)
	require.NoError(t, err, "If this fails, then there's no point running the rest of these tests")
	knownTypes = append(knownTypes, embNT)

	runTest := func(name, goString string, dtMod func(dt *DataType)) {
		t.Run(name, func(t *testing.T) {
			dt := newTestDataType()
			dt.Name = name
			dtMod(&dt) // Allows caller to modify dt make it more complicated
			nt, err := dt.AsType(knownTypes)
			assert.NoError(t, err)
			assert.Equal(t, name, nt.GoName())
			assert.Equal(t, name, nt.PlcName())
			assert.Equal(t, goString, nt.GoTypeString())
		})
	}

	runTest(
		"IncludeRegisteredType",
		"struct {\n\tVarName int16\n\tOtherVar RegisteredType `plc:\"otherVar\"`\n}",
		func(dt *DataType) {
			dt.Members = append(dt.Members, newTestMember("otherVar", "RegisteredType"))
		},
	)
	runTest(
		"TIMER",
		"struct {\n\tTimerVar TIMER `plc:\"timerVar\"`\n}",
		func(dt *DataType) {
			dt.Members = []Member{newTestMember("timerVar", "TIMER")}
		},
	)
	runTest(
		"EmbeddedTIMER",
		"struct {\n\tTIMER\n}",
		func(dt *DataType) {
			dt.Members = []Member{newTestMember("TIMER", "TIMER")}
		},
	)
}

func TestDataTypeAsNamedTypeError(t *testing.T) {
	_, err := newTestDataType().AsType(NewTypeList())
	require.NoError(t, err, "If this DataType isn't valid, these subtests are meaningless")

	runTestExpectingError := func(name string, dtMod func(dt *DataType)) {
		t.Run(name, func(t *testing.T) {
			dt := newTestDataType()
			dtMod(&dt) // Allows caller to modify dt to break something
			_, err := dt.AsType(NewTypeList())
			assert.Error(t, err)
		})
	}

	runTestExpectingError("NoFamily", func(dt *DataType) {
		dt.Family = -1
	})
	runTestExpectingError("MissingMemberName", func(dt *DataType) {
		dt.Members[0].Name = ""
	})
	runTestExpectingError("BadMemberType", func(dt *DataType) {
		dt.Members[0].DataType = "INVALID_TYPE"
	})
	runTestExpectingError("EmptyStruct", func(dt *DataType) {
		dt.Members = nil
	})
	runTestExpectingError("EmptyStructBit", func(dt *DataType) {
		dt.Members[0].DataType = "BIT"
	})
	runTestExpectingError("NoClass", func(dt *DataType) {
		dt.Class = -1
	})
	runTestExpectingError("StringBadRadix", func(dt *DataType) {
		dtToStringDT(dt)
		dt.Members[1].Radix = RadixDecimal
	})
	runTestExpectingError("StringExcessMembers", func(dt *DataType) {
		dtToStringDT(dt)
		dt.Members = append(dt.Members, newTestMember("otherVar", "BIT"))
	})
	runTestExpectingError("StringOneMember", func(dt *DataType) {
		dtToStringDT(dt)
		dt.Members = dt.Members[0:1]
	})
	runTestExpectingError("StringNoDimension", func(dt *DataType) {
		dtToStringDT(dt)
		dt.Members[1].Dimension = 0
	})
	runTestExpectingError("StringLenDimension", func(dt *DataType) {
		dtToStringDT(dt)
		dt.Members[0].Dimension = 1
	})
	runTestExpectingError("StringLenType", func(dt *DataType) {
		dtToStringDT(dt)
		dt.Members[0].DataType = "SINT"
	})
	runTestExpectingError("StringDataDataType", func(dt *DataType) {
		dtToStringDT(dt)
		dt.Members[0].DataType = "INT"
	})
}
