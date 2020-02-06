runondocker:# 运行在docker中
	docker-compose up -d
	@echo "docker-compose build complete ,current time is " `date`
runonraw:# 原始前台运行
	go build -v
	./server
runonrawbk:# 原始后台运行
	go build -v
	nohup ./server &
	