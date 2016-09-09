# 基于区域链的返利系统demo

## 背景
  返利系统主要关注社交场景中的分享/转载链和下单/下载时触发的的返利策略结算，基于hyperledger的区块链技术解决方案可以很好的满足这些需求场景。
## 名词解释
- hyperledger 超级帐本，由Linux基金会发起并管理，该项目的愿景是借助项目成员和开源社区的合力，制定一个开放的、跨行业、跨国界的区块链技术开源标准，打造可以跨行业的区块链解决方案
- fabric 超级帐本的开源项目名称
- chaincode 应用层代码，也叫智能合约
- ledger 帐本，由block链组成。 block由transaction组成。 
- transaction  由一个request来执行ledger上的一个function. function由chaincode实现
- peer 是一种在网络里负责执行一致性协议、确认交易和维护账本的计算机节点
- pbft  实用拜占庭容错算法, 用来解决一致性。
- mxcc mx chaincode

## 功能介绍
   以用户分享链接,用户点击链接下载app为产品背景。本demo提供以下功能： 商家充值&转帐&查询；用户分享/下载链维护；下载transaction触发智能合约实现返利策略；用户帐户查询；lastfour返利策略定义。

## 系统介绍
### hyperledger的架构
![](http://i.imgur.com/XjBaug7.png)
### mxcc 架构
![](http://i.imgur.com/my84NPV.png)
### 介绍
#### 商家模块
- 帐户充值
- 帐户查询
- 为返利链上的用户返利（转帐）
- 关联商品及反复策略
#### 用户模块
- 帐户余额查询
- 帐户提现
- 发起分享/转载transaction 
- 发起购买/下载transaction
#### 积分
- 商家充值来购买积分，也是积分的唯一生成源
- 返利策略触发时积分由商家转移到用户帐号
#### 结算
- 商家充值
- 用户提现
#### chaincode 智能合约
- 用户的购买/下载行为触发返利智能合约，执行相应返利策略
#### blockchain
- 记录系统所有行为（transaction）的帐本（数据库）

### demo 安装及演示
1.  fabric开发环境配置  
    参照官方文档： [http://hyperledger-fabric.readthedocs.io/en/latest/Setup/Chaincode-setup/](http://hyperledger-fabric.readthedocs.io/en/latest/Setup/Chaincode-setup/) , 确保在fabric/build/bin 目录下运行： peer node start --peer-chaincodedev 成功  （启动chaincode验证节点）
2.  返利chaincode安装与部署
     源代码见： [https://github.com/huilixiang/mxcc](https://github.com/huilixiang/mxcc)
     将chaincode_meixin目录copy到 fabric/examples/chaincode/go目录下
     运行go build成功后，CORE_CHAINCODE_ID_NAME=mxcc CORE_PEER_ADDRESS=0.0.0.0:7051 ./chaincode_meixin （部署chaincode）
3. demo脚本
   https://github.com/huilixiang/mxcc/blob/master/chaincode_meixin/demo.go
    