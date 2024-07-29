# new sql

# `NewSQL`?

- RDBMS가 지원하는 피처들 (SQL, 트랜잭션, 등) 을 다 지원하면서 분산스케일링이 되는 디비들
- `NewSQL`의 할아버지같은 DB가 GCP Cloud Spanner
    - GCP Cloud Spanner가 처음 등장(스패너 자체는 SQL은 읽기밖에 지원안해서 NewSQL은 아님)
    - Spanner 개발자들이 퇴사해서 만든게 [CockroachDB](https://www.cockroachlabs.com)이고
    - Spanner가 잘되어서 TiDB, FaunaDB, Vitess 등등이 나오고 이름이 NewSQL이라고 붙여졌음
- TiDB, FaunaDB는 유즈케이스를 제가 잘 모르는데 Spanner랑 CockroachDB는 여기저기서 많이 쓰고있고 Vitess는 유튜브에서 쓰인다
- 이게 좋긴한데 손으로 띄우게되면 DevOps 코스트가 많이 든다
- 그리고 스프링클라우드, TiDB클라우드, Cockroach클라우드 이런 서드파티 클라우드 쓰면 문제가 AWS환경이랑 연결하는게 잘 안되거나 심지어 일부가 아예 안되거나 이런상황이 좀 있는데, GCP 제품은 AWS-GCP연결은 그래도 (비교적) 잘 되는 편이라 AWS 사용중이라면 GCP 제품이 그나마 낫다
- Vercel도 서드파티클라우드라고 볼 수 있는데, vercel 서버랑 AWS private subnet이 연결이 안되어서 지금도 저희가 public한것들만 vercel에 띄우고있다
- 결제 정보를 Saas 서비스에 올리려면 고객 동의가 필요할 것으로 보여짐 (GCP 예외, 내부망이 연결되어있음)
