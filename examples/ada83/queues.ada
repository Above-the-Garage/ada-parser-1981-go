-- A generic bounded queue, with an instantiation and a client.
generic
   type Item is private;
   Size : in Positive := 16;
package Queues is
   type Queue is limited private;
   procedure Put (Q : in out Queue; X : in Item);
   procedure Get (Q : in out Queue; X : out Item);
   function Is_Empty (Q : Queue) return Boolean;
   Empty, Full : exception;
private
   type Buffer is array (1 .. Size) of Item;
   type Queue is
      record
         Data        : Buffer;
         Head, Count : Natural := 0;
      end record;
end Queues;

package body Queues is
   procedure Put (Q : in out Queue; X : in Item) is
   begin
      if Q.Count = Size then
         raise Full;
      end if;
      Q.Data ((Q.Head + Q.Count) mod Size + 1) := X;
      Q.Count := Q.Count + 1;
   end Put;

   procedure Get (Q : in out Queue; X : out Item) is
   begin
      if Is_Empty (Q) then
         raise Empty;
      end if;
      X := Q.Data (Q.Head + 1);
      Q.Head := (Q.Head + 1) mod Size;
      Q.Count := Q.Count - 1;
   end Get;

   function Is_Empty (Q : Queue) return Boolean is
   begin
      return Q.Count = 0;
   end Is_Empty;
end Queues;

with Queues, Text_IO;
use Text_IO;
procedure Queue_Demo is
   package Int_Queues is new Queues (Item => Integer, Size => 4);
   use Int_Queues;
   Q : Queue;
   N : Integer;
begin
   for I in reverse 1 .. 3 loop
      Put (Q, I * 10);
   end loop;
   while not Is_Empty (Q) loop
      Get (Q, N);
      Put_Line ("Got" & Integer'Image (N));
   end loop;
exception
   when Empty | Full =>
      Put_Line ("queue error");
   when others =>
      raise;
end Queue_Demo;
