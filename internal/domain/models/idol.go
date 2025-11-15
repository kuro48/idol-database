package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Idol struct {
	ID bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string `bson:"name" json:"name" binding:"required"`
    Group       string `bson:"group" json:"group"`
    Birthdate   string `bson:"birthdate" json:"birthdate"`
    Nationality string `bson:"nationality" json:"nationality"`
    ImageURL    string `bson:"image_url" json:"image_url"`
}

type CreateIdolRequest struct {
	Name        string `json:"name" binding:"required"`
	Group       string `json:"group"`
    Birthdate   string `json:"birthdate"`
    Nationality string `json:"nationality"`
    ImageURL    string `json:"image_url"`
}

type UpdateIdolRequest struct {
	Name        string `json:"name,omitempty"`
	Group       string `json:"group,omitempty"`
    Birthdate   string `json:"birthdate,omitempty"`
    Nationality string `json:"nationality,omitempty"`
    ImageURL    string `json:"image_url,omitempty"`
}