# starrocks_web
一个简单的项目，运营和监控starrocks（基于sr3.3.7版本）
包括sr的routine load,物化视图作业的查询，任务的启动和停止，界面化的操作更方便
 -routine load的基础作业信息，创建时间、作业停止时间、作业创建信息、导入数据量和错误信息等
 -物化视图的的基础作业信息、数据量
 -物化视图的依赖关系查询
 
基础效果:
<img width="1742" alt="image" src="https://github.com/user-attachments/assets/4b0d62b7-1822-4fa1-829b-fdc7e2a33749" />
<img width="1757" alt="image" src="https://github.com/user-attachments/assets/736d438e-f03e-4564-8456-af9a521f6d0c" />
<img width="1707" alt="image" src="https://github.com/user-attachments/assets/dbbf1111-ca04-496a-8122-9e8aded2cdcd" />


欢迎大家提建议和使用！😊

附上安装和启动步骤:
1. 服务器上安装了go环境，我的是go1.22.1，比这个版本高都行
2. 导入项目需要的依赖，所有的依赖都在go.sum文件里面
3. 修改main.go里面的main函数，把数据库用户名/密码、IP地址改成自己的就行;
4. 启动代码，nohup go run main.go &

启动后，本地可以通过http://localhost:6080/访问页面

如果想修改访问端口和访问地址，修改r.run()里面的ip地址和端口就行


