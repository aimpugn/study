package com.example.testing.controller;


import com.example.testing.AppConfig;
import com.example.testing.mapper.UserMapper;
import com.example.testing.model.User;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.BDDMockito;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.annotation.DirtiesContext;
import org.springframework.test.context.ContextConfiguration;
import org.springframework.test.context.junit.jupiter.SpringExtension;
import org.springframework.test.context.web.WebAppConfiguration;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

import javax.sql.DataSource;
import java.util.List;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@WebAppConfiguration
@ContextConfiguration(classes = AppConfig.class)
@ExtendWith(SpringExtension.class)
public class UserControllerTest {
    @Nested
    class UsingDb {
        @Autowired
        WebApplicationContext context;
        MockMvc mockMvc;
        @Autowired
        DataSource dataSource;

        @BeforeEach
        void setup() throws Exception {
            try (var conn = dataSource.getConnection();
                 var stmt = conn.createStatement()) {
                stmt.execute("CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100));");
                stmt.execute("DELETE FROM users");
                stmt.execute("INSERT INTO users(name) VALUES ('hello'), ('world'), ('Alice'), ('Bob');");
            }

            mockMvc = MockMvcBuilders.webAppContextSetup(context).build();
        }

        @Test
        @DirtiesContext(methodMode = DirtiesContext.MethodMode.AFTER_METHOD)
        void testGetUsersUsingDb() throws Exception {
            mockMvc.perform(get("/users").accept(MediaType.APPLICATION_JSON))
                .andExpect(status().isOk())
                .andExpect(result ->
                    System.out.println("DB 응답: " + result.getResponse().getContentAsString()));
        }
    }

    @Nested
    class UsingMock {
        @Autowired
        WebApplicationContext context;
        MockMvc mockMvc;
        @MockBean
        UserMapper userMapper;

        @BeforeEach
        void setup() {
            mockMvc = MockMvcBuilders.webAppContextSetup(context).build();
            BDDMockito.given(userMapper.findAll()).willReturn(List.of(
                new User(1, "a"),
                new User(2, "b"),
                new User(3, "c")
            ));
        }

        @Test
        @DirtiesContext(methodMode = DirtiesContext.MethodMode.AFTER_METHOD)
        void testGetUsersUsingMockito() throws Exception {
            mockMvc.perform(get("/users").accept(MediaType.APPLICATION_JSON))
                .andExpect(status().isOk())
                .andExpect(result ->
                    System.out.println("Mock 응답: " + result.getResponse().getContentAsString()));
        }
    }
}
