>>>工具说明
测试工具分配两部分，测试执行程序和配置文件。
配置文件在cfg目录下。可修改的是config.yaml,其他的配置文件只可以做协议初始值赋值不能修改其他的。
config.yaml中配置发包的目的地址和端口号。如果测试测试IP和Port非默认值请依次修改。


>>>协议测试说明
1.测试GTP-U协议。
	执行测试目录下gtpu，在打开的配置界面配置要发送的协议包的每一项。比如选择Message Type的G-PDU。
 	*其中Tunnel Status为3GPP较新的消息类型，wireshark等可能不识别，但不影响测试工具使用。
2.测试PFCP协议。
	执行测试目录下pfcp，在打开的配置界面配置要发送的协议包的每一项。
	支持增加PFCP的IE发送。
3.测试SCTP协议。
	先在命令行执行测试目录下的simple-sctp服务端。命令为：
		./simple-sctp -server -port 3868 -ip 127.0.0.1
	*注意port和ip与配置文件(config.yaml)中的对应

	执行测试目录下sctp，在打开的配置界面配置要发送的协议包的每一项。主要配置TSN、PPID和Chunk data。
	支持多chunk配置。
4.测试NGAP协议。
	同SCTP协议，先在命令行执行测试目录下的simple-sctp服务端。命令为：
		./simple-sctp -server -port 3868 -ip 127.0.0.1

	执行测试目录下ngap，在打开的配置界面配置要发送的协议包的每一项。
	*PPID(Payload Protocol Identifier) 必须为60才会识别为NGAP协议。
