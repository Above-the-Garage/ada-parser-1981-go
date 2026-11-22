function GCD (A, B : Natural) return Natural is
   X : Natural := A;
   Y : Natural := B;
   T : Natural;
begin
   while Y /= 0 loop
      T := X mod Y;
      X := Y;
      Y := T;
   end loop;
   return X;
end GCD;
