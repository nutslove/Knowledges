- 参考URL
  - https://www.programiz.com/python-programming/decorator

- 以下の場合、`func`はdecoratorを呼び出している関数`after_login`を指す
  - 同じ色にしているものは同じもので、合わせる必要がある
  - まず`inner`関数が実行される  
    ![decorator](image/decorator.jpg)


~~~python
def make_pretty(func):
    # define the inner function 
    def inner():
        # add some additional behavior to decorated function
        print("I got decorated")

        # call original function
        func()
    # return the inner function
    return inner

# define ordinary function
def ordinary():
    print("I am ordinary")
    
# decorate the ordinary function
decorated_func = make_pretty(ordinary)

# call the decorated function
decorated_func()
~~~

- decorator例
~~~python
def smart_divide(func):
    def inner(a, b):
        print("I am going to divide", a, "and", b)
        if b == 0:
            print("Whoops! cannot divide")
            return

        return func(a, b)
    return inner

@smart_divide
def divide(a, b):
    print(a/b)

divide(2,5)

divide(2,0)
~~~