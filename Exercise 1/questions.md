Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> Concurrency is multithreding on same hardware, but the threads dont run simultaneously, while paraellism runs n sepearet hardware at the same time.

What is the difference between a *race condition* and a *data race*? 
> Race condition occurs when two resources tries to accsess same resource at the same time. Similarly, Data race is when two threads are trying to access data at the same address in memory. Examplewise if one thread is attempting to write to the address, while another wants to read the same data.
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> Determines when to swap between threads and which threads


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> When you want mulltiple "programs" to finish at the same time. 

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> Smaller seperated parts of threads, divide and conquer. Uses cooperative scheduling, which means the program determines when to switch threads.

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> Makes it way harder to find bugs, can make some code way smoother and readible.

What do you think is best - *shared variables* or *message passing*?
> Message passing is better because global variables are scary, no idea when they change value. 
