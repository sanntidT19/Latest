package optimalization

import (
	"fmt"
)
const (

)

//Return IPorder

func Optimalization_init(m Master) {
	for {
		ord := <-ExOptimalChans.OptimizationTriggerChan
		
		slaveIPs := m.SLaveIp
		workloadVector := make([]int, N_ELEV)
		var distance int

		for i:=0; i < len(SlaveIp); i++ {
			if distance = ord.Order.Floor - m.Statelist[slaveIPs[i]]; distance > 0{
				workloadVector[i] += distance
				for j := 0; j < N_FLOORS; j++{
					for k := 0; k < 2; k++{
						if m.ExternalList[slaveIPs[i]][j][k]{ //Order in floor
							if k != ord.Direction{
								if j < ord.Order.Floor{
									workloadVector[i] += 2*(ord.Order.Floor - j)
								}else{
									workloadVector[i] += ord.Order.Floor - j
								}
							}
							if k == ord.Direction{
								if j < ord.Order.Floor{
									workloadVector[i] += ord.Order.Floor - j
								}else{
									workloadVector[i] += 2*(j - ord.Order.Floor)
								}
							}
						}
					}

				}
			}else{
				workloadVector[i] -= distance
				for j := 0; j < N_FLOORS; j++{
					for k := 0; k < 2; k++{
						if m.ExternalList[slaveIPs[i]][j][k]{ //Order in floor
							if k != ord.Direction{
								if j > ord.Order.Floor{
									workloadVector[i] += 2*(j -ord.Order.Floor)
								}else{
									workloadVector[i] += ord.Order.Floor - j
								}
							}
							if k == ord.Direction{
								if j > ord.Order.Floor{
									workloadVector[i] += j - ord.Order.Floor 
								}else{
									workloadVector[i] += 2*(ord.Order.Floor - j)
								}

							}
						}
					}

				}
			}

		}
		//Chooses elevator with lead workload
		NrMinWorkloadElevator := 0
		tempLoad :=	9999
		for i:=0; i<len(slaveIPs); i++ {
			if tempLoad >= workLoadVector[i]{
				NrMinWorkloadElevator = i
				tempLoad = workLoadVector[i]
			}
		}
		ord.Ip = slaveIPs[NrMinWorkloadElevator]
		ExOptimalChans.OptimizationReturnChan <- ord
	}
}