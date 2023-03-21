package database

// // DBUpdateCloudInstance update cloud instance
// func DBUpdateCloudInstance(inputObj *CloudInstance) {
// 	var obj CloudInstance
// 	DB.Model(CloudInstance{}).Where(CloudInstance{IPAddress: inputObj.IPAddress, InputName: inputObj.InputName}).First(&obj)
// 	if obj.ID != 0 {
// 		// @TODO: change target to old
// 		DB.Model(&obj).Updates(inputObj)
// 		inputObj.ID = obj.ID

// 	}
// 	DB.Create(&inputObj)
// }

// // GetRunningClouds get running cloud instance
// func GetRunningClouds() []CloudInstance {
// 	var objs []CloudInstance
// 	DB.Model(CloudInstance{}).Where(CloudInstance{Status: "running"}).Find(&objs)
// 	return objs
// }
