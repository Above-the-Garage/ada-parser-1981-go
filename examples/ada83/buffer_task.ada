-- A one-slot buffer task and its rendezvous.
procedure Producer_Consumer is
   task Buffer is
      entry Write (X : in Integer);
      entry Read (X : out Integer);
   end Buffer;

   task type Worker;
   W : Worker;

   task body Buffer is
      Value : Integer;
      Full  : Boolean := False;
   begin
      loop
         select
            when not Full =>
               accept Write (X : in Integer) do
                  Value := X;
               end Write;
               Full := True;
         or
            when Full =>
               accept Read (X : out Integer) do
                  X := Value;
               end Read;
               Full := False;
         or
            delay 10.0;
            exit;
         end select;
      end loop;
   end Buffer;

   task body Worker is
      N : Integer;
   begin
      Buffer.Write (42);
      Buffer.Read (N);
   end Worker;

   function "+" (L, R : Boolean) return Boolean is separate;

   for Integer'Size use 32;
begin
   null;
end Producer_Consumer;
