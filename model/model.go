package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type (
	Image_Details struct {
		ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
		Name         string             `json:"name,omitempty" bson:"_name,omitempty"`
		DateOfUpload primitive.DateTime `json:"date_of_upload,omitempty" bson:"img_date_of_upload,omitempty"`
		Nickname     string             `json:"nickname,omitempty" bson:"_nickname,omitempty"`
		DownloadLink string             `json:"download_link,omitempty" bson:"img_download_link,omitempty"`
	}
)
