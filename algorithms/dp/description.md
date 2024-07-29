# Dynamic programming

- [Dynamic programming](#dynamic-programming)
  - [설명](#설명)

## 설명

Dynamic programming is a method for solving problems by breaking them down into smaller subproblems and solving each subproblem only once, caching the result and reusing it whenever needed. This approach allows us to reduce the time complexity of the solution.

Here's a step-by-step explanation on how to use dynamic programming for finding the most efficient way of hundreds of thousands of combinations:

1. Identify the subproblems:
   - Break down the problem into smaller, overlapping subproblems that can be solved individually.
2. Define the state:
   - Determine the state of the subproblem, which describes the inputs and outputs required to solve it.
3. Define the state transition function:
   - Create a function that takes the current state and returns the next state.
   - This function should be used to iteratively solve the subproblems.
4. Create a table to store the results:
   - Create a table to store the results of each subproblem.
   - The table should be indexed by the state of the subproblem.
5. Populate the table:
   - Starting from the subproblems with the simplest states, compute the results of each subproblem and store them in the table.
6. Use the table to find the final result:
   - Use the results stored in the table to find the final result of the original problem.
7. Optimize the solution:
   - Optimize the solution by reducing the size of the table, using memoization, or using other techniques to reduce the time and space complexity of the solution.

Dynamic programming can be a powerful approach for solving problems with many possible combinations, especially if the subproblems have overlapping solutions. However, it is important to carefully consider the problem and the state transition function in order to ensure that the approach is correct and efficient.
