package com.example.testing.transaction;

import com.example.testing.AppConfig;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.test.context.ContextConfiguration;
import org.springframework.test.context.junit.jupiter.SpringExtension;
import org.springframework.test.context.web.WebAppConfiguration;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

import javax.sql.DataSource;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@WebAppConfiguration
@ContextConfiguration(classes = AppConfig.class)
@ExtendWith(SpringExtension.class)
public class NestedTransaction {
    @Autowired
    WebApplicationContext wac;

    @Autowired
    DataSource dataSource;


    private MockMvc mockMvc;
    private JdbcTemplate jdbcTemplate;
    private ObjectMapper objectMapper = new ObjectMapper();

    @BeforeEach
    void setup() {
        mockMvc = MockMvcBuilders.webAppContextSetup(wac).build();
        jdbcTemplate = new JdbcTemplate(dataSource);

//        stmt.execute("CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100));");
        var testTableScheme = """
            CREATE TABLE IF NOT EXISTS test_table (id VARCHAR(36) PRIMARY KEY, val VARCHAR(100));
            """.trim();
        jdbcTemplate.execute(testTableScheme);
    }

    @Test
    void unexpectedRollbackException() throws Exception {
        // Controller 통해 outer 호출 (정상 실행)
        var result = mockMvc.perform(get("/tx/unexpectedRollbackException")
                .contentType(MediaType.APPLICATION_JSON))
            .andExpect(status().isOk()).andReturn();

        var content = result.getResponse().getContentAsString();
        System.out.println("result: " + content);
    }

    @Test
    void resolveUnexpectedRollbackException() throws Exception {
        // Controller 통해 outer 호출 (정상 실행)
        var result = mockMvc.perform(get("/tx/resolveUnexpectedRollbackException")
                .contentType(MediaType.APPLICATION_JSON))
            .andExpect(status().isOk()).andReturn();

        var content = result.getResponse().getContentAsString();
        System.out.println("result: " + content);
    }
}