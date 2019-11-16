package main

import  (
   "fmt"
   storkcli "stork-cli"
)

func main() {
   fmt.Println("STORK-CLI")

   var client *storkcli.APIClient;

   // var ctx context.Context;

   cfg := storkcli.NewConfiguration();
   cfg.BasePath = "http://localhost:8888";
   client = storkcli.NewAPIClient(cfg)

   version, httpResp, err := client.GeneralApi.GetVersion(nil);

   fmt.Println("### version=");
   fmt.Println(version);
   fmt.Println("### httpResp=");
   fmt.Println(httpResp);

   if err != nil {
      fmt.Println("### err=");
      fmt.Println(err);
   }
}

