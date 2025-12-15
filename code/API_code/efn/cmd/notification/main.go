/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

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
	"go.uber.org/zap"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/reciever/notification"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/config"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/event"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/logger"
)

func main() {
	conf := config.GetConf()
	log := logger.Get()

	server, err := event.NewReceiver(conf.API)
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := database.NewMongoDB(conf.Database)
	if err != nil {
		log.With(zap.Error(err), zap.String("Mongo URI", conf.Database.Uri), zap.String("DB Name", conf.Database.Name)).
			Fatal("failed to connect to mongo Database")
	}

	handler := notification.NewHandler(db, conf.HTTP)

	log.With(zap.String("address", conf.API.Address)).Info("Starting notification server")
	err = server.Start(handler)
	if err != nil {
		log.With(zap.Error(err)).Fatal("Failed to start event receiver")
	}
}
