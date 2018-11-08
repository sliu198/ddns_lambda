package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "log"
    "math"
    "net"
    "strconv"
    "strings"
    "time"
    "./github.com/aws/aws-sdk-go/aws/session"
    "./github.com/aws/aws-sdk-go/service/route53"
    "encoding/base64"
    "os"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/aws"
)

var allowedNames = make(map[string]bool)
var key []byte
var hostedZoneId string

var sess, _ = session.NewSession()
var r53 = route53.New(sess)

type Event struct{
    Params Params `json:"queryStringParameters"`
}

type Params struct {
    Ip string `json:"ip"`
    Name string `json:"name"`
    Timestamp string `json:"timestamp"`
    Hmac string `json:"hmac"`
}

type Response struct{
    StatusCode int `json:"statusCode"`
    Body string `json:"body"`
}

func init() {
    env := os.Environ()
    for _,entry := range env {
        comps := strings.Split(entry,"=")
        switch comps[0] {
        case "ALLOWED_NAMES":
            var a []string
            json.Unmarshal([]byte(comps[1]),&a)
            for _,n := range a {
                allowedNames[n] = true
            }
        case "SHARED_SECRET":
            key,_ = base64.RawURLEncoding.DecodeString(comps[1])
        case "HOSTED_ZONE_ID":
            hostedZoneId = comps[1]
        }
    }
}

func Handler(event Event) (Response, error) {
    if key == nil {
        log.Println("[ERROR] missing environment variable SHARED_KEY")
        return Response{500,""}, nil
    }

    if hostedZoneId == "" {
        log.Println("[ERROR] missing environment variable HOSTED_ZONE_ID")
        return Response{500,""}, nil
    }

    params := event.Params

    ipString := params.Ip
    if ipString == "" {
        return Response{400,"missing parameter ip"}, nil
    }

    name := params.Name
    if name == "" {
        return Response{400,"missing parameter name"}, nil
    }

    timestampString := params.Timestamp
    if timestampString == "" {
        return Response{400,"missing parameter timestamp"}, nil
    }

    authString := params.Hmac
    if authString == "" {
        return Response{400,"missing parameter hmac"}, nil
    }

    timestamp, err := strconv.ParseInt(timestampString,10,64)
    if err != nil {
        return Response{400,"invalid timestamp format"}, nil
    }

    ip := net.ParseIP(ipString)
    if ip == nil {
        return Response{400,"invalid ip format"}, nil
    }

    auth,err := base64.RawURLEncoding.DecodeString(authString)
    if err != nil {
        return Response{400,"invalid hmac format"}, nil
    }

    h := hmac.New(sha256.New,key)
    h.Write([]byte(fmt.Sprintf("ip=%v&name=%v&timestamp=%v", ipString, name, timestampString)))
    if !hmac.Equal(h.Sum(nil), auth) || math.Abs(float64(time.Now().Unix() - timestamp)) > 30 {
        return Response{401, ""}, nil
    }

    if !allowedNames[name] {
        return Response{403, ""}, nil
    }

    _, err = r53.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
       ChangeBatch: &route53.ChangeBatch{
           Changes: []*route53.Change{
               {
                   Action: aws.String("UPSERT"),
                   ResourceRecordSet: &route53.ResourceRecordSet{
                       Name: aws.String(name),
                       ResourceRecords: []*route53.ResourceRecord{
                           {
                               Value: aws.String(ipString),
                           },
                       },
                       TTL:  aws.Int64(300),
                       Type: aws.String("A"),
                   },
               },
           },
       },
       HostedZoneId: aws.String(hostedZoneId),
    })

    if err != nil {
        log.Printf("[ERROR] error updating; name=%v; ip=%v; hostedZoneId=%v; error=%v", name, ipString, hostedZoneId, err)
        return Response{500, ""}, nil
    }

    return Response{204, ""}, nil
}

func main() {
    lambda.Start(Handler)
}
