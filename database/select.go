package database

import (
    "gorm.io/gorm/clause"
)

// jsoniter "github.com/json-iterator/go"

type WorkspaceData struct {
}

func GetWorkspaces() []Target {
    var objs []Target
    DB.Preload(clause.Associations).Preload("Reports").Find(&objs).Order("created_at desc")
    return objs
}

func GetScans() []Scan {
    var objs []Scan
    DB.Preload(clause.Associations).Preload("Targets").Find(&objs).Order("created_at desc")
    return objs
}

func GetClouds() []CloudInstance {
    var objs []CloudInstance
    DB.Preload(clause.Associations).Preload("Targets").Find(&objs).Order("created_at desc")
    return objs
}
