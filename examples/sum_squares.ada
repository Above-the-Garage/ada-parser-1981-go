-- Sum of the first N squares.
procedure Sum_Squares is
   N     : Integer := 10;
   Total : Integer := 0;
begin
   for I in 1 .. N loop
      Total := Total + I * I;
   end loop;
end Sum_Squares;
