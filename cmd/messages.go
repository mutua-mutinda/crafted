package cmd

import ( 

	"github.com/spf13/cobra" 
	// "github.com/vonmutinda/crafted/api/messages"
	// "github.com/vonmutinda/crafted/config"
)


var receiveCmd = &cobra.Command{
	Use: "consume",
	Aliases: []string{"consume", "c"},
	Short: "Consume messages send via RabbitMQ",
	Long: `Listen to queue...`,
	Run: func(cmd *cobra.Command, args []string){
		receive()
	},
}

func init(){ 
	rootCmd.AddCommand(receiveCmd)
}

func receive(){  
	// config.Load() 
	// messages.Consume()

}