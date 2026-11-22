with Text_IO; use Text_IO;
procedure Shapes is
   type Kind is (Circle, Square, Triangle);
   type Real is digits 6;
   subtype Small is Integer range 0 .. 99;
   Pi : constant := 3.14159_26536;

   type Shape (K : Kind := Circle) is
      record
         X, Y : Real := 0.0;
         case K is
            when Circle =>
               Radius : Real;
            when Square =>
               Side : Real;
            when Triangle =>
               A, B, C : Real;
         end case;
      end record;

   type Shape_Ptr is access Shape;
   type Shape_List is array (Positive range <>) of Shape_Ptr;

   function Area (S : Shape) return Real is
      P : Real;
   begin
      case S.K is
         when Circle =>
            return Real (Pi) * S.Radius ** 2;
         when Square =>
            return S.Side * S.Side;
         when Triangle =>
            P := (S.A + S.B + S.C) / 2.0;
            return P * (P - S.A) * (P - S.B) * (P - S.C);
      end case;
   end Area;

   List  : Shape_List (1 .. 3) :=
     (new Shape'(K => Circle, X => 0.0, Y => 0.0, Radius => 1.0),
      new Shape'(Square, 1.0, 1.0, 2.0),
      others => null);
   Total : Real := 0.0;
   Count : Small := 0;
begin
   Outer :
   for I in List'Range loop
      exit Outer when List (I) = null;
      Total := Total + Area (List (I).all);
      Count := Count + 1;
   end loop Outer;
   if Count > 0 and then Total /= 0.0 then
      Put_Line ("Average area computed");
   elsif Count = 0 or else Total = 0.0 then
      Put_Line ("Nothing to do");
   else
      null;
   end if;
   declare
      Name : constant String := "shapes";
      Ch   : Character := 'x';
   begin
      if Ch in 'a' .. 'z' and Name (1 .. 2) = "sh" then
         <<Done>> null;
      end if;
   end;
end Shapes;
