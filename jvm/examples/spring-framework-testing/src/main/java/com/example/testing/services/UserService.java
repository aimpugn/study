package com.example.testing.services;

import com.example.testing.mapper.UserMapper;
import com.example.testing.model.User;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class UserService {
    private final UserMapper userMapper;

    public UserService(UserMapper userMapper) {
        this.userMapper = userMapper;
    }

    public List<User> getUsers() {
        var result = userMapper.findAll();
        System.out.println("mapper result: " + result);
        return result;
    }
}
