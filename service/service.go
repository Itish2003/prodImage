package service

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"prodImage/controller"
	"prodImage/model"

	// "text/template"

	"time"

	// "github.com/fogleman/primitive"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var s3Client *s3.S3

func init() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-north-1"),
	}))
	s3Client = s3.New(sess)
}

func generateUniqueName(name string) string {
	return time.Now().Format("2006") + "_" + name
}

func UpdateImage(c *gin.Context) {
	filter := bson.D{bson.E{Key: "_name", Value: c.Param("name")}}

	var updateFields map[string]interface{}
	if err := c.ShouldBindJSON(&updateFields); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build the update document based on user input
	update := bson.D{{Key: "$set", Value: updateFields}}

	// Perform the update operation
	result, err := controller.Collection.UpdateOne(c, filter, update)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if any document was modified
	if result.ModifiedCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No matching document found for update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document updated successfully"})
}

func DeleteImage(c *gin.Context) {
	filter := bson.D{bson.E{Key: "_name", Value: c.Param("name")}}
	result, err := controller.Collection.DeleteOne(c, filter)
	if err != nil {
		log.Fatal(err)
	}
	if result.DeletedCount != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No matching document found for delete"})
		return
	}

}

func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to get file: %v", err)
		log.Printf("FormFile error: %v", err)
		return
	}

	filename := generateUniqueName(file.Filename)

	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to open uploaded file: %v", err)
		log.Printf("Open file error: %v", err)
		return
	}
	defer fileContent.Close()

	// Upload the file to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("imgeffects"), // replace with your S3 bucket name
		Key:    aws.String(filename),
		Body:   fileContent,
		ACL:    aws.String(s3.ObjectCannedACLPublicRead), // Adjust ACL if needed
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to upload file to S3: %v", err)
		log.Printf("S3 PutObject error: %v", err)
		if s3Err, ok := err.(awserr.Error); ok {
			log.Printf("S3 Error Code: %s, Message: %s", s3Err.Code(), s3Err.Message())
		}
		return
	}

	// Download the file from S3 to a temporary location
	tmpFile, err := os.CreateTemp("", filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to create temporary file: %v", err)
		log.Printf("CreateTemp error: %v", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("imgeffects"),
		Key:    aws.String(filename),
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to download file from S3: %v", err)
		log.Printf("S3 GetObject error: %v", err)
		if s3Err, ok := err.(awserr.Error); ok {
			log.Printf("S3 Error Code: %s, Message: %s", s3Err.Code(), s3Err.Message())
		}
		return
	}
	defer result.Body.Close()

	_, err = io.Copy(tmpFile, result.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to save temporary file: %v", err)
		log.Printf("Save temporary file error: %v", err)
		return
	}

	// Process the image file using the primitive command
	cmd := exec.Command("C:/Users/itish/go/bin/primitive", "-i", tmpFile.Name(), "-o", "static/processed/"+filename, "-n", "100")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error running primitive command: %v", err)
		log.Printf("Primitive command error: %v", err)
		return
	}

	// Open the processed file
	processedFilePath := "static/processed/" + filename
	processedFile, err := os.Open("static/processed/" + filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to open processed file: %v", err)
		log.Printf("Open processed file error: %v", err)
		return
	}
	defer processedFile.Close()

	// Upload the processed file to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("imgeffects"), // replace with your S3 bucket name
		Key:    aws.String("processed/" + filename),
		Body:   processedFile,
		ACL:    aws.String(s3.ObjectCannedACLPublicRead), // Adjust ACL if needed
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to upload processed file to S3: %v", err)
		log.Printf("S3 PutObject error: %v", err)
		if s3Err, ok := err.(awserr.Error); ok {
			log.Printf("S3 Error Code: %s, Message: %s", s3Err.Code(), s3Err.Message())
		}
		return
	}

	// Remove the processed file after uploading
	if err := os.Remove(processedFilePath); err != nil {
		c.String(http.StatusInternalServerError, "Unable to delete processed file: %v", err)
		log.Printf("Delete processed file error: %v", err)
		return
	}

	nickname := c.PostForm("nickname")

	// Generate S3 file URL
	fileURL := "https://imgeffects.s3.eu-north-1.amazonaws.com/processed/" + filename

	imgDetails := &model.Image_Details{
		Name:         filename,
		DateOfUpload: primitive.NewDateTimeFromTime(time.Now()),
		Nickname:     nickname,
		DownloadLink: fileURL,
	}

	_, err = controller.Collection.InsertOne(c, imgDetails)
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to insert image details into database: %v", err)
		log.Printf("MongoDB InsertOne error: %v", err)
		return
	}

	c.HTML(http.StatusOK, "result.html", gin.H{
		"Name":         imgDetails.Name,
		"DownloadLink": imgDetails.DownloadLink,
	})
}

func GetImage(c *gin.Context) {
	var img *model.Image_Details
	filter := bson.D{bson.E{Key: "_name", Value: c.Param("name")}}
	err := controller.Collection.FindOne(c, filter).Decode(&img)
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"File ID":        img.ID.Hex(),
		"Filename":       img.Name,
		"Nickname":       img.Nickname,
		"Date Of Upload": img.DateOfUpload,
		"Download Link":  img.DownloadLink,
	})
}

func GetAllImage(c *gin.Context) {
	var imgs []*model.Image_Details
	cursor, err := controller.Collection.Find(c, bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(c)

	for cursor.Next(c) {
		var img *model.Image_Details
		err := cursor.Decode(&img)
		if err != nil {
			log.Fatal(err)
		}
		imgs = append(imgs, img)
	}
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
	if len(imgs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "no images found",
		})
		return
	}
	c.JSON(http.StatusOK, imgs)
}

func DownloadImage(c *gin.Context) {
	filename := c.Param("name")
	bucketName := "your-bucket-name" // replace with your S3 bucket name

	// Retrieve the file from S3
	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to retrieve file from S3")
		return
	}
	defer result.Body.Close()

	// Set headers to indicate file download
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", *result.ContentType)
	c.DataFromReader(http.StatusOK, *result.ContentLength, *result.ContentType, result.Body, nil)
}

func GetHome(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}
