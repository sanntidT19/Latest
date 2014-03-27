package optimal

import (
	. "chansnstructs"
	"fmt"
)

//Return IPOrd

func Optimalization_init(m *Master) {
	for {
		optOrd := <-ExOptimalChans.OptimizationTriggerChan
		slaveIps := m.SlaveIp
		workloadVector := make([]int, N_ELEV)
		var distance int

		for i := 0; i < len(slaveIps); i++ {
			if slaveIps[i] == nil {
				workloadVector[i] = 9999
			}
			if slaveIps[i] != nil {
				if distance = optOrd.Ord.Floor - m.Statelist[i].Sta.CurrentFloor; distance > 0 {
					workloadVector[i] += distance
					for j := 0; j < N_FLOORS; j++ {
						for k := 0; k < 2; k++ {
							if m.ExternalList[i][j][k] { //Ord in floor
								if k != optOrd.Ord.ButtonType {
									if j < optOrd.Ord.Floor {
										workloadVector[i] += 2 * (optOrd.Ord.Floor - j)
									} else {
										workloadVector[i] += optOrd.Ord.Floor - j
									}
								}
								if k == optOrd.Ord.ButtonType {
									if j < optOrd.Ord.Floor {
										workloadVector[i] += optOrd.Ord.Floor - j
									} else {
										workloadVector[i] += 2 * (j - optOrd.Ord.Floor)
									}
								}
							}
						}

					}
				} else {
					workloadVector[i] -= distance
					for j := 0; j < N_FLOORS; j++ {
						for k := 0; k < 2; k++ {
							if m.ExternalList[i][j][k] { //Ord in floor
								if k != optOrd.Ord.ButtonType {
									if j > optOrd.Ord.Floor {
										workloadVector[i] += 2 * (j - optOrd.Ord.Floor)
									} else {
										workloadVector[i] += optOrd.Ord.Floor - j
									}
								}
								if k == optOrd.Ord.ButtonType {
									if j > optOrd.Ord.Floor {
										workloadVector[i] += j - optOrd.Ord.Floor
									} else {
										workloadVector[i] += 2 * (optOrd.Ord.Floor - j)
									}

								}
							}
						}

					}
				}
			}

		}
		//Chooses elevator with lead workload
		NrMinWorkloadElevator := 0
		tempLoad := 9999
		for i := 0; i < len(slaveIps); i++ {
			if tempLoad >= workloadVector[i] {
				NrMinWorkloadElevator = i
				tempLoad = workloadVector[i]
			}
		}
		fmt.Println(workloadVector)
		optOrd.Ip = slaveIps[NrMinWorkloadElevator]
		fmt.Println("optimalization finished with this one:   ", optOrd)
		ExOptimalChans.OptimizationReturnChan <- optOrd
	}
}
