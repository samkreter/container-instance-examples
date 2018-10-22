import random
import string
import sys
from collections import deque
from config import queueConfig, azure_context, ACI_CONFIG
from azure.servicebus import ServiceBusService, Message, Queue
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.containerinstance import ContainerInstanceManagementClient
from azure.mgmt.containerinstance.models import (ContainerGroup, Container, ContainerPort, Port, IpAddress, EnvironmentVariable,
                                                 ResourceRequirements, ResourceRequests, ContainerGroupNetworkProtocol, OperatingSystemTypes)

client = ContainerInstanceManagementClient(azure_context.credentials, azure_context.subscription_id)


bus_service = ServiceBusService(
    service_namespace = queueConfig['service_namespace'],
    shared_access_key_name = queueConfig['saskey_name'],
    shared_access_key_value = queueConfig['saskey_value'])

IMAGE = "pskreter/worker-container:latest"


def main():
    sys.stdout.write("Starting Work Cycle...\n")  # same as print
    sys.stdout.flush()
    
    while True:
        try:
            msg = bus_service.receive_queue_message(queueConfig['queue_name'], peek_lock=False)
            if msg.body is not None:
                work = msg.body.decode("utf-8")

                container_name = generate_container_name()
                env_vars = [EnvironmentVariable(name = "MESSAGE", value = work), EnvironmentVariable(name = "CONTAINER_NAME", value = container_name)]
                
                sys.stdout.write("Creating container: " + container_name + " with work: " + work + '\n')  # same as print
                sys.stdout.flush()
                
                create_container_group(ACI_CONFIG['resourceGroup'], container_name, ACI_CONFIG['location'], IMAGE, env_vars)
                
        except KeyboardInterrupt:
            pass


def generate_container_name():
    return ''.join(random.choice(string.ascii_lowercase + string.digits) for _ in range(7))


def create_container_group(resource_group_name, name, location, image, env_vars):

    # setup default values
    port = 80
    container_resource_requirements = None
    command = None

    # set memory and cpu
    container_resource_requests = ResourceRequests(memory_in_gb = 3.5, cpu = 2)
    container_resource_requirements = ResourceRequirements(requests = container_resource_requests)

    container = Container(name = name,
                        image = image,
                        resources = container_resource_requirements,
                        command = command,
                        ports = [ContainerPort(port=port)],
                        environment_variables = env_vars)

    # defaults for container group
    cgroup_os_type = OperatingSystemTypes.linux
    cgroup_ip_address = IpAddress(ports = [Port(protocol=ContainerGroupNetworkProtocol.tcp, port = port)])
    image_registry_credentials = None

    cgroup = ContainerGroup(location = location,
                        containers = [container],
                        os_type = cgroup_os_type,
                        ip_address = cgroup_ip_address,
                        image_registry_credentials = image_registry_credentials)

    client.container_groups.create_or_update(resource_group_name, name, cgroup)


if __name__ == '__main__':
    main()