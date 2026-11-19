# chair

A booking app for a small hair salon, called chair.

The salon offers services. Each has a name, how many minutes it takes, a price in euros and a short description. The salon publishes the times it is open as slots: each slot is a start time for one service. Clients book a free slot; a slot that someone booked, or that has already started, is not free.

There is no login yet. A visitor picks which client they are from a list, or adds themselves with their name, email and an optional phone number, and the app remembers the choice.

Pages:
- The home page: a hero that says "Your chair is waiting." and "Book a cut, a color or a shave in a few taps." with a "Book a time" button, then every service.
- Pick who you are: every client, tap one to become them and go to booking. Under it, "New here?" with a button to add yourself.
- Add a client: a form for name, email and phone.
- Services: every service as a card with its name, minutes and price, each linking to its page.
- A service's page: its name, minutes, price and description, with a "Book" button.
- Book: every free slot from now on, soonest first, with its time and service, under "Pick a time". Tapping one books it for the visitor. It says "No free times right now." when there are none.
- My bookings: the visitor's bookings from now on, soonest first: the slot's time, when it was booked and the note. Each can be cancelled. It says "Nothing booked yet." when there are none.
- Me: the visitor's name, email and phone.
- Forms to add a service and to add a slot.

Navigation: a bar at the top with Services, Book, My bookings and Me.

Look: a calm, upscale hair salon: warm stone and deep olive, editorial serif headings.

Booking fails with a message when the slot is already taken or has already started.
