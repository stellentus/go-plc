package l5x

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testFilePath = "test.L5X"

var exampleController = Controller{
	Use:                      "Target",
	Name:                     "EXAMPLE_FACTORY",
	ProcessorType:            "1769-L33ER",
	MajorRev:                 30,
	MinorRev:                 1,
	TimeSlice:                20,
	ShareUnusedTimeSlice:     1,
	ProjectCreationDate:      rsLogixTime(time.Date(2004, 4, 30, 6, 12, 44, 0, time.UTC)),
	LastModifiedDate:         rsLogixTime(time.Date(2020, 10, 5, 16, 36, 36, 0, time.UTC)),
	SFCExecutionControl:      "CurrentActive",
	SFCRestartPosition:       "MostRecent",
	SFCLastScan:              "DontScan",
	ProjectSN:                "16#6096_bdb0",
	PassThroughConfiguration: "EnabledWithAppend",
	DownloadProjDocs:         true,
	DownloadProjProps:        true,
	RedundancyInfo: RedundancyInfo{
		IOMemoryPadPercentage:  90,
		DataTablePadPercentage: 50,
	},
	Security:     Security{ChangesToDetect: "16#ffff_ffff_ffff_ffff"},
	SecurityInfo: struct{}{},
	DataTypes: []DataType{
		DataType{
			Name: "dow",
			Members: []Member{
				Member{
					Name:     "DayOW",
					DataType: "INT",
				},
				Member{
					Name:     "Month",
					DataType: "DINT",
				},
				Member{
					Name:      "MonthCode",
					DataType:  "DINT",
					Dimension: 13,
				},
				Member{
					Name:     "DayOW1",
					DataType: "REAL",
					Radix:    RadixFloat,
				},
			},
		},
		DataType{
			Name: "big_data_type",
			Members: []Member{
				Member{
					Name:     "XprivateX_cleaning_c0",
					DataType: "SINT",
					Hidden:   true,
				},
				Member{
					Name:        "CLEAN_RATE",
					DataType:    "BIT",
					BitNumber:   0,
					Target:      "XprivateX_cleaning_c0",
					Description: Description{Cdata: "\nRate at which cleaning occurs\n"},
				},
				Member{
					Name:        "CLEAN_COMPLEXITY",
					DataType:    "BIT",
					BitNumber:   1,
					Target:      "XprivateX_cleaning_c0",
					Description: Description{Cdata: "\nComplexity of cleaning job\n"},
				},
				Member{
					Name:        "FUN_FACTOR",
					DataType:    "BIT",
					BitNumber:   2,
					Target:      "XprivateX_cleaning_c0",
					Description: Description{Cdata: "\nMeasure of how fun the job is\n"},
				},
				Member{
					Name:        "PRODUCT_COST",
					DataType:    "BIT",
					BitNumber:   3,
					Target:      "XprivateX_cleaning_c0",
					Description: Description{Cdata: "\nAccumulated cost of all products used\n"},
				},
				Member{
					Name:        "AJIBSH_35",
					DataType:    "BIT",
					BitNumber:   4,
					Target:      "XprivateX_cleaning_c0",
					Description: Description{Cdata: "\nSettings for the AJIBSH_35\n"},
				},
				Member{
					Name:        "CLEAN_MODE",
					DataType:    "INT",
					Description: Description{Cdata: "\nCleaning Mode 0=Constant, 1=Cyclic, 2=Up/Down, 3=None, 4=Lasers\n"},
				},
				Member{
					Name:     "XprivateX_cleaning_c7",
					DataType: "SINT",
					Hidden:   true,
				},
				Member{
					Name:        "VALVE_ENABLE",
					DataType:    "BIT",
					BitNumber:   0,
					Target:      "XprivateX_cleaning_c7",
					Description: Description{Cdata: "\nValve Enable\n"},
				},
				Member{
					Name:        "TIGER_SUBSYSTEM",
					DataType:    "BIT",
					BitNumber:   1,
					Target:      "XprivateX_cleaning_c7",
					Description: Description{Cdata: "\nEnable the tiger system\n"},
				},
				Member{
					Name:        "REVERSE_TIME_BUTTON",
					DataType:    "BIT",
					BitNumber:   2,
					Target:      "XprivateX_cleaning_c7",
					Description: Description{Cdata: "\nToggle status of the reverse time button\n"},
				},
			},
		},
		DataType{
			Name: "datas_for_eating",
			Members: []Member{
				Member{
					Name:     "XprivateX_cleaning_c0",
					DataType: "SINT",
					Hidden:   true,
				},
				Member{
					Name:        "DEMAND",
					DataType:    "BIT",
					BitNumber:   0,
					Target:      "XprivateX_cleaning_c0",
					Description: Description{Cdata: "\nHow much eating is demanded?\n"},
				},
				Member{
					Name:        "FOOD_TIMER",
					DataType:    "TIMER",
					Radix:       RadixNullType,
					Description: Description{Cdata: "\nTimer tracking the food consumption\n"},
				},
				Member{
					Name:        "MEAL_PREP_TIMER",
					DataType:    "TIMER",
					Radix:       RadixNullType,
					Description: Description{Cdata: "\nTimer for amount of meal prep time\n"},
				},
				Member{
					Name:        "BHAIG29GI",
					DataType:    "TIMER",
					Radix:       RadixNullType,
					Description: Description{Cdata: "\nTimer\nfor the\nBHAIG29GI\n"},
				},
				Member{
					Name:        "COUNTDOWN_TO_DESSERT",
					DataType:    "TIMER",
					Radix:       RadixNullType,
					Description: Description{Cdata: "\nIndicates when dessert is done\n"},
				},
				Member{
					Name:        "STEPS_REQUIRED",
					DataType:    "INT",
					Description: Description{Cdata: "\nSteps required to burn enough calories\n"},
				},
			},
		},
	},
	Modules: []Module{
		Module{
			Name:            "Local",
			CatalogNumber:   "1769-L33ER",
			Vendor:          1,
			ProductType:     14,
			ProductCode:     107,
			Major:           30,
			Minor:           1,
			ParentModule:    "Local",
			ParentModPortId: 1,
			Inhibited:       false,
			MajorFault:      true,
			Ports: []Port{
				Port{
					Id:      1,
					Address: "0",
					Type:    PortTypeCompact,
					Bus: struct {
						Size int `xml:",attr,omitempty"`
					}{Size: 99},
				},
				Port{
					Id:      2,
					Address: "192.168.1.170",
				},
			},
		},
		Module{
			Name:            "AI1",
			CatalogNumber:   "1769-IF4C/A",
			Vendor:          1,
			ProductType:     10,
			ProductCode:     12,
			Major:           1,
			Minor:           1,
			ParentModule:    "Local",
			ParentModPortId: 1,
			EKey:            EKeyState_s{State: EKeyStateCompatibleModule},
			Ports: []Port{
				Port{
					Id:       1,
					Address:  "1",
					Type:     PortTypeCompact,
					Upstream: true,
				},
			},
			Communications: Communications{
				ConfigTag: ConfigTag{
					ConfigSize: 200,
					Data: []Data{
						Data{
							Format: DataFormatL5K,
							L5K:    "\n[The analog input AI1 has some config which belongs here]\n",
						},
						Data{
							L5K: "\n\n",
							Structure: Structure{
								DataType: "AB:1769_IF4:C:0",
								DataValueMember: []DataValueMember{
									DataValueMember{Name: "RealTimeSample", DataType: "INT", Value: "0"},
									DataValueMember{Name: "TimestampEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch00Filter", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch00AlarmLatchEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch00AlarmEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch00En", DataType: "BOOL", Value: "1"},
									DataValueMember{Name: "Ch00RangeType", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch00DataFormat", DataType: "SINT", Value: "1"},
									DataValueMember{Name: "Ch00HAlarmLimit", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch00LAlarmLimit", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch00AlarmDeadband", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch01Filter", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch01AlarmLatchEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch01AlarmEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch01En", DataType: "BOOL", Value: "1"},
									DataValueMember{Name: "Ch01RangeType", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch01DataFormat", DataType: "SINT", Value: "1"},
									DataValueMember{Name: "Ch01HAlarmLimit", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch01LAlarmLimit", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch01AlarmDeadband", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch02Filter", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch02AlarmLatchEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch02AlarmEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch02En", DataType: "BOOL", Value: "1"},
									DataValueMember{Name: "Ch02RangeType", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch02DataFormat", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch02HAlarmLimit", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch02LAlarmLimit", DataType: "INT", Value: "1"},
									DataValueMember{Name: "Ch02AlarmDeadband", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch03Filter", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch03AlarmLatchEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch03AlarmEn", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch03En", DataType: "BOOL", Value: "1"},
									DataValueMember{Name: "Ch03RangeType", DataType: "SINT", Value: "0"},
									DataValueMember{Name: "Ch03DataFormat", DataType: "SINT", Value: "4"},
									DataValueMember{Name: "Ch03HAlarmLimit", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch03LAlarmLimit", DataType: "INT", Value: "0"},
									DataValueMember{Name: "Ch03AlarmDeadband", DataType: "INT", Value: "4"},
								},
							},
						},
					},
				},
				Connections: []Connection{Connection{
					Name: "Output",
					RPI:  64000,
					Type: IOTypeOutput,
					InputTag: IOTag{
						Comments: []Comment{
							Comment{Operand: ".CH00DATA", Cdata: "\nDX-15 Dance Cyclant\n"},
							Comment{Operand: ".CH01DATA", Cdata: "\nGBY-5 Sound Level\n"},
							Comment{Operand: ".CH02DATA", Cdata: "\nGG-46 Hurricaner\n"},
							Comment{Operand: ".CH03DATA", Cdata: "\nOH-00 Firestorm\n"},
						},
						Data: []Data{Data{
							L5K: "\n\n",
							Structure: Structure{
								DataType: "AB:1769_IF4:I:0",
								DataValueMember: []DataValueMember{
									DataValueMember{Name: "Fault", DataType: "DINT", Radix: RadixBinary, Value: "2#0000_0000_0000_0000_0000_0000_0000_0000"},
									DataValueMember{Name: "Ch00Data", DataType: "INT", Value: "3457"},
									DataValueMember{Name: "Ch01Data", DataType: "INT", Value: "5234"},
									DataValueMember{Name: "Ch02Data", DataType: "INT", Value: "2722"},
									DataValueMember{Name: "Ch03Data", DataType: "INT", Value: "2622"},
									DataValueMember{Name: "Timestamp", DataType: "INT", Value: "4"},
									DataValueMember{Name: "Ch00Status", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch01Status", DataType: "BOOL", Value: "1"},
									DataValueMember{Name: "Ch02Status", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch03Status", DataType: "BOOL", Value: "1"},
									DataValueMember{Name: "Ch00Overrange", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch00Underrange", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch00HAlarm", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch00LAlarm", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch01Overrange", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch01Underrange", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch01HAlarm", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch01LAlarm", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch02Overrange", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch02Underrange", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch02HAlarm", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch02LAlarm", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch03Overrange", DataType: "BOOL", Value: "1"},
									DataValueMember{Name: "Ch03Underrange", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch03HAlarm", DataType: "BOOL", Value: "0"},
									DataValueMember{Name: "Ch03LAlarm", DataType: "BOOL", Value: "0"},
								},
							},
						}},
					},
					OutputTag: IOTag{
						Data: []Data{
							Data{
								Format: DataFormatL5K,
								L5K:    "\n[0,0]\n",
							},
							Data{
								L5K: "\n\n",
								Structure: Structure{
									DataType: "AB:1769_IF4:O:0",
									DataValueMember: []DataValueMember{
										DataValueMember{Name: "Ch00HAlarmUnlatch", DataType: "BOOL", Value: "0"},
										DataValueMember{Name: "Ch00LAlarmUnlatch", DataType: "BOOL", Value: "0"},
										DataValueMember{Name: "Ch01HAlarmUnlatch", DataType: "BOOL", Value: "0"},
										DataValueMember{Name: "Ch01LAlarmUnlatch", DataType: "BOOL", Value: "0"},
										DataValueMember{Name: "Ch02HAlarmUnlatch", DataType: "BOOL", Value: "0"},
										DataValueMember{Name: "Ch02LAlarmUnlatch", DataType: "BOOL", Value: "0"},
										DataValueMember{Name: "Ch03HAlarmUnlatch", DataType: "BOOL", Value: "0"},
										DataValueMember{Name: "Ch03LAlarmUnlatch", DataType: "BOOL", Value: "0"},
									},
								},
							},
						},
					},
				}},
			},
			ExtendedProperties: ExtendedProperties{Public: struct {
				ConfigID int
				CatNum   string
			}{ConfigID: 100, CatNum: "1769-IF4C"}},
		},
	},
	AddOnInstrDefs: []AddOnInstrDef{AddOnInstrDef{
		Name:              "EVENT_TOT",
		Revision:          "1.0",
		RevisionExtension: "0",
		Vendor:            "Cool Stuff",
		CreatedDate:       iso8601Time(time.Date(1987, 5, 12, 15, 21, 28, 170000000, time.UTC)),
		CreatedBy:         "GESMAM\\414206527",
		EditedDate:        iso8601Time(time.Date(1987, 5, 12, 15, 25, 19, 828000000, time.UTC)),
		EditedBy:          "GESMAM\\414032557",
		SoftwareRevision:  "v12.69",
		Description:       Description{Cdata: "\nLife Excitement\n"},
		Parameters: []Parameter{
			Parameter{
				Name:           "EnableIn",
				DataType:       "BOOL",
				ExternalAccess: ExternalAccessReadOnly,
				Description:    Description{Cdata: "\nEnable Input - System Defined Parameter\n"},
			},
			Parameter{Name: "EnableOut",
				DataType:       "BOOL",
				Usage:          IOTypeOutput,
				ExternalAccess: ExternalAccessReadOnly,
				Description:    Description{Cdata: "\nEnable Output - System Defined Parameter\n"},
			},
			Parameter{Name: "AlarmSP",
				DataType:    "INT",
				Required:    true,
				Visible:     true,
				Description: Description{Cdata: "\nExcitement High Alarm Setpoint\n"},
				DefaultData: []Data{
					Data{
						Format: DataFormatL5K,
						L5K:    "\n0\n",
					},
					Data{
						L5K: "\n\n",
						DataValue: DataValueMember{
							DataType: "INT",
							Radix:    RadixDecimal,
							Value:    "0",
						},
					},
				},
			},
		},
		LocalTags: []LocalTag{LocalTag{
			Name:        "EarplugArray",
			DataType:    "INT",
			Dimensions:  20,
			Description: Description{Cdata: "\nEarplug Array\n"},
			DefaultData: []Data{
				Data{
					Format: DataFormatL5K,
					L5K:    "\n[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]\n",
				},
				Data{
					L5K: "\n\n",
					Array: Array{
						DataType:   "INT",
						Dimensions: 20,
						Radix:      RadixDecimal,
						Elements: []Element{
							Element{Index: 0, Value: "0"},
							Element{Index: 1, Value: "0"},
							Element{Index: 2, Value: "0"},
							Element{Index: 3, Value: "0"},
							Element{Index: 4, Value: "0"},
							Element{Index: 5, Value: "0"},
							Element{Index: 6, Value: "0"},
							Element{Index: 7, Value: "0"},
							Element{Index: 8, Value: "0"},
							Element{Index: 9, Value: "0"},
							Element{Index: 10, Value: "0"},
							Element{Index: 11, Value: "0"},
							Element{Index: 12, Value: "0"},
							Element{Index: 13, Value: "0"},
							Element{Index: 14, Value: "0"},
							Element{Index: 15, Value: "0"},
							Element{Index: 16, Value: "0"},
							Element{Index: 17, Value: "0"},
							Element{Index: 18, Value: "0"},
							Element{Index: 19, Value: "0"},
						},
					},
				},
			},
		}},
		Routines: []Routine{Routine{
			Name: "Logic",
			RLLContent: struct {
				Rungs []Rung `xml:"Rung"`
			}{Rungs: []Rung{
				Rung{
					Number: 0,
					Text:   Description{Cdata: "\n[MOV(1,23) DO(92,TheThing)];\n"},
				},
				Rung{
					Number: 1,
					Text:   Description{Cdata: "\n[ACT(My) ,ACT(Oh) ACT(My) ][DOTHING(0,5,12,26,74)];\n"},
				},
			},
			},
		},
		},
	}},
	Tags: []Tag{
		Tag{
			Name:     "ALARM_P",
			DataType: "alarm_info",
			Data: []Data{
				Data{
					Format: DataFormatL5K,
					L5K:    "\n[[0,0,0],0]\n",
				},
				Data{
					L5K: "\n\n",
					Structure: Structure{
						DataType: "alarm_info",
						StructureMember: []DataValueMember{
							{Name: "PRE", DataType: "DINT", Value: "0"},
							{Name: "ACC", DataType: "DINT", Value: "0"},
							{Name: "EN", DataType: "BOOL", Value: "1"},
							{Name: "TT", DataType: "BOOL", Value: "0"},
						},
						DataValueMember: []DataValueMember{
							{Name: "ALM_ACTIVE", DataType: "BOOL", Value: "0"},
						},
					},
				},
			},
		},
		Tag{
			Name:        "INFO_ABOUT",
			DataType:    "INT",
			Dimensions:  2,
			Description: Description{Cdata: "\nInfo data\n"},
			Data: []Data{
				Data{
					Format: DataFormatL5K,
					L5K:    "\n[-2925,1952]\n",
				},
				Data{
					L5K: "\n\n",
					Array: Array{
						DataType:   "INT",
						Dimensions: 2,
						Radix:      RadixDecimal,
						Elements: []Element{
							Element{Index: 0, Value: "-2925"},
							Element{Index: 1, Value: "1952"},
						},
					},
				},
			},
		},
		Tag{
			Name:        "BIGGD",
			DataType:    "big_data_type",
			Description: Description{Cdata: "\nBig Data Lots\n"},
			Data: []Data{
				Data{
					Format: DataFormatL5K,
					L5K:    "\n[1,7,3]\n",
				},
				Data{
					L5K: "\n\n",
					Structure: Structure{
						DataType: "big_data_type",
						DataValueMember: []DataValueMember{
							DataValueMember{
								Name:     "CLEAN_RATE",
								DataType: "BOOL",
								Value:    "0",
							},
							DataValueMember{
								Name:     "CLEAN_COMPLEXITY",
								DataType: "BOOL",
								Value:    "1",
							},
							DataValueMember{
								Name:     "FUN_FACTOR",
								DataType: "BOOL",
								Value:    "0",
							},
							DataValueMember{
								Name:     "PRODUCT_COST",
								DataType: "BOOL",
								Value:    "0",
							},
							DataValueMember{
								Name:     "AJIBSH_35",
								DataType: "BOOL",
								Value:    "0",
							},
							DataValueMember{
								Name:     "CLEAN_MODE",
								DataType: "INT",
								Radix:    RadixDecimal,
								Value:    "7",
							},
							DataValueMember{
								Name:     "VALVE_ENABLE",
								DataType: "BOOL",
								Value:    "1",
							},
							DataValueMember{
								Name:     "TIGER_SUBSYSTEM",
								DataType: "BOOL",
								Value:    "1",
							},
							DataValueMember{
								Name:     "REVERSE_TIME_BUTTON",
								DataType: "BOOL",
								Value:    "0",
							},
						},
					},
				},
			},
		},
	},
	Programs: []Program{Program{
		Name:            "DANCER",
		MainRoutineName: "MainRoutine",
		Tags: []Tag{Tag{
			Name:        "DOW",
			DataType:    "dow",
			Description: Description{Cdata: "\nDay of the Week\n"},
			Data: []Data{
				Data{
					Format: DataFormatL5K,
					L5K:    "\n[3,12,[0,5,2,7,5,0,2,5,1,4,6,3,4],6.2]\n",
				},
				Data{
					L5K: "\n\n",
					Structure: Structure{
						DataType: "dow",
						DataValueMember: []DataValueMember{
							DataValueMember{Name: "DayOW", DataType: "INT", Value: "3"},
							DataValueMember{Name: "Month", DataType: "DINT", Value: "12"},
							DataValueMember{Name: "DayOW1", DataType: "REAL", Radix: RadixFloat, Value: "6.2"},
						},
						ArrayMember: Array{Name: "MonthCode",
							DataType:   "DINT",
							Dimensions: 13,
							Radix:      RadixDecimal,
							Elements: []Element{
								Element{Index: 0, Value: "0"},
								Element{Index: 1, Value: "5"},
								Element{Index: 2, Value: "2"},
								Element{Index: 3, Value: "7"},
								Element{Index: 4, Value: "5"},
								Element{Index: 5, Value: "0"},
								Element{Index: 6, Value: "2"},
								Element{Index: 7, Value: "5"},
								Element{Index: 8, Value: "1"},
								Element{Index: 9, Value: "4"},
								Element{Index: 10, Value: "6"},
								Element{Index: 11, Value: "3"},
								Element{Index: 12, Value: "4"},
							},
						},
					},
				},
			},
		}},
		Routines: []Routine{Routine{
			Name:        "INITIATE_DANCE_SEQUENCE",
			Description: Description{Cdata: "\nCode to Initiate the dance sequence\n"},
			RLLContent: struct {
				Rungs []Rung `xml:"Rung"`
			}{Rungs: []Rung{
				Rung{
					Number:  0,
					Comment: Description{Cdata: "\n===================================================================================================================================================================================\nINITIATE DANCE SEQUENCE - ver 0.1\n\nROUTINE FUNCTION\n1. Turn on lights.\n2. Turn on sprinklers.\n3. Crank music up to 3.\n\nCODE DETAILS\nRung 0 - Read all sensor data and save it.\nRung 1 - Be confused\nRung 2 - Admit complete ignorance of rungs\nRung 3 - Phone someone until it rings\nRung 4 - It has run\n===================================================================================================================================================================================\n"},
					Text:    Description{Cdata: "\nTHING(Fancy code goes here);\n"},
				},
				Rung{Number: 1,
					Text: Description{Cdata: "\nBeep() Beep() Bloop();\n"},
				},
				Rung{Number: 2,
					Text: Description{Cdata: "\n[Computer(?,?,?);\n"},
				},
				Rung{Number: 3,
					Text: Description{Cdata: "\n[Succeed();\n"},
				},
				Rung{Number: 4,
					Text: Description{Cdata: "\nDONE[]];\n"},
				},
			}},
		}},
	}},
	Tasks: []Task{Task{
		Name:     "MainTask",
		Priority: 10,
		Watchdog: 500,
		ScheduledPrograms: []struct {
			Name string `xml:",attr"`
		}{struct {
			Name string `xml:",attr"`
		}{Name: "DANCER"}},
	}},
	Trends: []Trend{Trend{
		Name:             "tacos",
		SamplePeriod:     500,
		NumberOfCaptures: 1,
		CaptureSizeType:  "Samples",
		CaptureSize:      340,
		StartTriggerType: "No Trigger",
		StopTriggerType:  "No Trigger",
		TrendxVersion:    "8.1",
		Template:         "255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255\n\n 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255\n\n 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255\n\n 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255 255\n",
		Pens: []Pen{Pen{
			Name:    "pen_name",
			Color:   "16#00ff_aa33",
			Visible: true,
			Type:    "Analog",
			Width:   1,
			Max:     100,
		}},
	}},
	TimeSynchronize: TimeSynchronize{Priority1: 64, Priority2: 64, PTPEnable: false},
	EthernetPorts: []EthernetPort{
		EthernetPort{Port: 1, Label: 1, PortEnabled: true},
		EthernetPort{Port: 2, Label: 2, PortEnabled: true},
	},
	EthernetNetwork: EthernetNetwork{
		SupervisorModeEnabled: true,
		SupervisorPrecedence:  5,
		BeaconInterval:        200,
		BeaconTimeout:         1864,
		VLANID:                15,
	},
}

var exampleRslogixContent = RSLogix5000Content{
	XMLName:          xml.Name{Local: "RSLogix5000Content"},
	SchemaRevision:   1,
	SoftwareRevision: 30,
	TargetName:       "EXAMPLE_FACTORY",
	TargetType:       "Controller",
	ExportDate:       rsLogixTime(time.Date(2020, 12, 10, 12, 45, 25, 0, time.UTC)),
	ExportOptions:    stringSlice{[]string{"NoRawData", "L5KData", "DecoratedData", "ForceProtectedEncoding", "AllProjDocTrans"}},
	Controller:       exampleController,
}

func TestXmlFromFile(t *testing.T) {
	l5x, err := NewFromFile(testFilePath)
	require.NoError(t, err)
	require.Equal(t, &exampleRslogixContent, l5x)
}

func TestXmlMarshall(t *testing.T) {
	t.Skip()
	require.Fail(t, "Marshal tests aren't implemented yet")
}
