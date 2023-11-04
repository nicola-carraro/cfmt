#define CRT_SECURE_NO_WARNINGS

#include "stdio.h"

int main(void)
{
    char *name = "Niki";

    char letter = 'a';

    char letter2 = '\'';

    float f = 55.0f;

    double d = .4;

    int num = 007;

    float e = 123.456e-67;

    float e1 = 123e+86;

    printf("%d\n", num);

    if (d <= 1)
    {
        return -1;
    }
    return 0;
}