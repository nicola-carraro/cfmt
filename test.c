#ifndef __TEST_C
#define __TEST_C

#define CRT_SECURE_NO_WARNINGS


#define MACRO(num, str) {\
            printf("%d", num);\
            printf(" is");\
            printf(" %s number", str);\
            printf("\n");\
           } 
#define C8_LOG_ERROR(msg) {}

#define MACRO2(num, str) {printf("%d", num);printf(" is");printf(" %s number", str);printf("\n");} 
/*
              A multiline comment bla bla



  bla bla bla*/
#include "stdio.h"
#include "stdlib.h"

// Some comment
typedef struct {
    int bar;     char *baz;}Foo;

// A comment before struct declaration
    typedef struct {
    int bar;// A comment after a struct declaration
       
    
         char *baz;}Bar;

  int buzz(){return -1;}  

int main(void)
{

    struct {
    int bar;     char *baz;};
/*
              Another multiline commen


              
  bla bla bla
  
  
  */

    Foo z = {0};

    Foo zz = {123, "123"  ,  };

    Foo yy = {123, "123"  , // Another comment in a weird place
     };

    Foo zzz = {
        123,
         "123"
    };

    z.bar = 3;

    int //For some reason we put a comment between type and identifier
    goo;

    Foo *aa = &z;

    aa->bar =3;

    buzz();
    char *name = "Niki";

    char letter = 'a';

    char letter2 = '\'';

    float f = 55.0f;

    int a = (1)*(4);

    int b = 3&4;

    double d = .4;

    int i = b++;

    i = --b;

    i = i++ == i--;

    i = (i)++ == (i)--;

    i = ((i)++) == ((i)++);

    i = ++i == i--;

    int h = ++b;

    int num = 007;

    float e = 123.456e-67;

    float e1 = 123e+86;

    printf("%d\n", num);

    for (int i=0;i<3;i++)
    {
        printf("%d\n", i);
    }

    if (d <= 1)
    {
        return -1;
    }

    return 0;
}

#endif