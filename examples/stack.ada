package Stack is
   procedure Push (X : in Integer);
   function Pop return Integer;
   Overflow, Underflow : exception;
end Stack;

package body Stack is
   Max   : constant := 100;
   type Table is array (1 .. Max) of Integer;
   Space : Table;
   Index : Integer range 0 .. Max := 0;

   procedure Push (X : in Integer) is
   begin
      if Index >= Max then
         raise Overflow;
      end if;
      Index := Index + 1;
      Space (Index) := X;
   end Push;

   function Pop return Integer is
   begin
      if Index = 0 then
         raise Underflow;
      end if;
      Index := Index - 1;
      return Space (Index + 1);
   end Pop;
end Stack;
