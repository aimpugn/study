package lab;

import com.zaxxer.hikari.HikariConfig;
import com.zaxxer.hikari.HikariDataSource;
import javax.sql.DataSource;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.datasource.DataSourceTransactionManager;
import org.springframework.transaction.annotation.EnableTransactionManagement;

/**
 * 실험 5의 Spring 설정입니다. Spring Boot 없이 Spring Framework만 씁니다. Boot의 자동 설정이 고르는 것(HikariCP,
 * {@code DataSourceTransactionManager} 계열, 프록시 기반 {@code @Transactional})을 손으로 같은 꼴로 세워, 화면에서 볼
 * 부품을 설정 한 장에 모두 드러내려는 것입니다.
 *
 * <p>접속 자리는 {@code -Dlab.url}로, 풀 크기는 {@code -Dlab.pool}로 받습니다(지휘가 시나리오의 pool 줄로 정함).
 */
@Configuration
@EnableTransactionManagement
public class LabConfig {

    /**
     * 연결 풀을 만듭니다. 풀이 처음 채울 연결 수를 최대 크기와 같게 두어, 장면이 시작될 때 연결이 모두 열려 있게 합니다.
     * 연결을 기다리는 시간은 2초로 줄였습니다(기본 30초). 풀이 1개일 때 REQUIRES_NEW가 연결을 기다리다 끝나는 장면을
     * 짧게 보려는 것입니다.
     *
     * @return 닫을 때 연결을 모두 닫는 풀
     * @throws IllegalStateException {@code -Dlab.url}이 없을 때
     */
    @Bean(destroyMethod = "close")
    public HikariDataSource dataSource() {
        String url = System.getProperty("lab.url");
        if (url == null) {
            throw new IllegalStateException("-Dlab.url이 필요합니다");
        }
        int pool = Integer.getInteger("lab.pool", 2);
        HikariConfig c = new HikariConfig();
        c.setPoolName("lab");
        c.setJdbcUrl(url);
        c.setUsername("postgres");
        c.setMaximumPoolSize(pool);
        c.setMinimumIdle(pool);
        c.setConnectionTimeout(2000);
        c.setValidationTimeout(1000);
        return new HikariDataSource(c);
    }

    /** JDBC 연결 하나로 트랜잭션을 여닫는 Spring의 트랜잭션 관리자입니다. */
    @Bean
    public DataSourceTransactionManager transactionManager(DataSource dataSource) {
        return new DataSourceTransactionManager(dataSource);
    }

    /** 지금 스레드에 묶인 트랜잭션의 연결이 있으면 그 연결로 SQL을 보내는 도우미입니다. */
    @Bean
    public JdbcTemplate jdbcTemplate(DataSource dataSource) {
        return new JdbcTemplate(dataSource);
    }

    @Bean
    public Inner inner(JdbcTemplate jdbc) {
        return new Inner(jdbc);
    }

    @Bean
    public Outer outer(JdbcTemplate jdbc, Inner inner) {
        return new Outer(jdbc, inner);
    }
}
