  val txShardingLayer: TaskLayer[TxSharding] = {
    val entityConfigLayer  = ZLayer.succeed(entityConfig).fresh
    val shardManagerClient = entityConfigLayer >>> ShardManagerClient.local.fresh

    val sharding =
      (entityConfigLayer ++
        Pods.noop ++
        shardManagerClient ++
        Serialization.javaSerialization ++
        scopedStorage.fresh) >>> Sharding.live.fresh

    ZLayer.scoped {
      sharding.memoize.map(layer =>
        ZLayer.make[TxSharding](
          entityConfigLayer,
          layer,
          ZLayer.fromFunction(TxSharding.apply(_: EntityConfig, _: Sharding)),
          ZLayer.succeed(GrpcConfig.default),
          GrpcShardingService.live.fresh,
        ),
      )
    }.flatten
  }