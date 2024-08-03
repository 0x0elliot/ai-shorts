package util

// import (
//     "context"
//     "fmt"
//     "io"
//     "log"
//     "os"
//     "path/filepath"
//     "sort"
//     "strings"

//     "cloud.google.com/go/storage"
//     ffmpeg "github.com/u2takey/ffmpeg-go"
//     "google.golang.org/api/iterator"
// )

// func StitchVideo(ctx context.Context, client *storage.Client, bucketName, folderPath string, sentences []string) error {
//     bucket := client.Bucket(bucketName)

//     // Download images
//     imageFiles, err := downloadObjects(ctx, bucket, folderPath, "image_")
//     if err != nil {
//         return fmt.Errorf("error downloading images: %v", err)
//     }

//     // Download audio
//     audioFile, err := downloadObject(ctx, bucket, filepath.Join(folderPath, "audio", "full_audio.mp3"), "full_audio.mp3")
//     if err != nil {
//         return fmt.Errorf("error downloading audio: %v", err)
//     }

//     // Sort image files
//     sort.Strings(imageFiles)

//     // Create temporary directory for processing
//     tempDir, err := os.MkdirTemp("", "video_processing")
//     if err != nil {
//         return fmt.Errorf("error creating temp dir: %v", err)
//     }
//     defer os.RemoveAll(tempDir)

//     // Process each image with its corresponding sentence
//     var inputs []ffmpeg.Input
//     for i, imagePath := range imageFiles {
//         if i >= len(sentences) {
//             break
//         }
//         sentence := sentences[i]
        
//         // Add image with text overlay
//         input := ffmpeg.Input(imagePath).
//             Filter("scale", ffmpeg.Args{"1080:1920"}).  // Instagram reel dimensions
//             DrawText(ffmpeg.Args{
//                 "fontfile=/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/Roboto-Regular.ttf",
//                 "fontsize=24",
//                 "fontcolor=white",
//                 "box=1",
//                 "boxcolor=black@0.5",
//                 "boxborderw=5",
//                 "x=(w-tw)/2",  // Center horizontally
//                 "y=h-th-20",   // Near bottom
//                 "text=" + sentence,
//             })
        
//         inputs = append(inputs, input)
//     }

//     // Concatenate all images
//     video := ffmpeg.Concat(inputs...).
//         Filter("fps", ffmpeg.Args{"30"}).  // Set framerate
//         Output(filepath.Join(tempDir, "video.mp4")).
//         OverWriteOutput().
//         ErrorToStdOut()

//     // Run FFmpeg command to create video
//     err = video.Run()
//     if err != nil {
//         return fmt.Errorf("error creating video: %v", err)
//     }

//     // Add audio to video
//     output := ffmpeg.Input(filepath.Join(tempDir, "video.mp4")).
//         Input(audioFile).
//         Output("/Users/aditya/Documents/OSS/zappush/shortpro/backend/public/video.mp4").
//         OverWriteOutput().
//         ErrorToStdOut()

//     err = output.Run()
//     if err != nil {
//         return fmt.Errorf("error adding audio to video: %v", err)
//     }

//     return nil
// }

// func downloadObjects(ctx context.Context, bucket *storage.BucketHandle, prefix, startsWith string) ([]string, error) {
//     var files []string
//     it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})
//     for {
//         attrs, err := it.Next()
//         if err == iterator.Done {
//             break
//         }
//         if err != nil {
//             return nil, err
//         }
//         if strings.HasPrefix(filepath.Base(attrs.Name), startsWith) {
//             localPath := filepath.Base(attrs.Name)
//             if _, err := downloadObject(ctx, bucket, attrs.Name, localPath); err != nil {
//                 return nil, err
//             }
//             files = append(files, localPath)
//         }
//     }
//     return files, nil
// }

// func downloadObject(ctx context.Context, bucket *storage.BucketHandle, object, destFile string) (string, error) {
//     src, err := bucket.Object(object).NewReader(ctx)
//     if err != nil {
//         return "", err
//     }
//     defer src.Close()

//     dst, err := os.Create(destFile)
//     if err != nil {
//         return "", err
//     }
//     defer dst.Close()

//     _, err = io.Copy(dst, src)
//     if err != nil {
//         return "", err
//     }

//     return destFile, nil
// }
