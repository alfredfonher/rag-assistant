# Hexagonal Architecture (Software)

The hexagonal architecture, also known as ports and adapters architecture, is an architectural style used in software design. It aims at creating loosely coupled application components that can be easily connected to their software environment through ports and adapters. This makes components exchangeable at any level and facilitates test automation.

## Origin

The hexagonal architecture was invented by Alistair Cockburn to avoid known structural pitfalls in object-oriented software design, such as undesired dependencies between layers and contamination of user interface code with business logic. In 2005, Cockburn renamed it "Ports and Adapters". In April 2024, Cockburn published a comprehensive book on the subject, coauthored with Juan Manuel Garrido de Paz.

The term "hexagonal" comes from the graphical conventions that show the application component as a hexagonal cell. The purpose was not to suggest that there would be six borders/ports, but to leave enough space to represent the different interfaces needed between the component and the external world.

## Principle

The hexagonal architecture divides a system into several loosely-coupled interchangeable components, such as the application core, the database, the user interface, test scripts, and interfaces with other systems. This approach is an alternative to the traditional layered architecture.

Each component is connected to the others through exposed "ports". Communication through these ports follows a protocol depending on their purpose. Ports and protocols define an abstract API that can be implemented by any suitable technical method (e.g., method invocation in an object-oriented language, remote procedure calls, or web services).

### Ports
The granularity of ports and their number is not constrained:
- A single port could be sufficient (e.g., simple service consumer)
- Typically, there are ports for event sources (UI, automatic feeding), notifications, database, and administration
- In extreme cases, there could be a different port for every use case

### Adapters
Adapters are the glue between components and the outside world. They tailor exchanges between the external world and ports representing the application's internal requirements. Multiple adapters can exist for one port (e.g., data can be provided through GUI, CLI, automated data source, or test scripts).

## Criticism

The term "hexagonal" implies six parts to the concept, whereas there are only four key areas. The term's usage comes from graphical conventions showing the application component as a hexagonal cell.

According to Martin Fowler, the hexagonal architecture has the benefit of using similarities between presentation layer and data source layer to create symmetric components, but with the drawback of hiding the inherent asymmetry between a service provider and a service consumer.

## Evolution

Some authors consider the hexagonal architecture to be at the origin of the microservices architecture.

## Variants

### Onion Architecture
Proposed by Jeffrey Palermo in 2008, similar to hexagonal architecture: it also externalizes infrastructure with interfaces to ensure loose coupling. It further decomposes the application core into concentric rings using inversion of control.

### Clean Architecture
Proposed by Robert C. Martin in 2012, combining principles of hexagonal, onion, and other variants. It provides additional detail levels presented as concentric rings, isolating adapters and interfaces in outer rings while keeping use cases and entities in inner rings. It uses dependency inversion with the strict rule that dependencies only exist from outer to inner rings.
