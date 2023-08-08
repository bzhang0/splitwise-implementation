package Splitwise;

import java.util.HashSet;
import java.util.Iterator;
import java.util.Set;

public class Group implements Comparable<Group> {
    private final int groupID; // guaranteed to be unique
    private String name;
    private String description;
    private Set<Person> members;

    public Group(int groupID, String name, String description) {
        this.groupID = groupID;
        this.name = name;
        this.description = description;
        this.members = new HashSet<>();
    }

    public int getGroupID() {
        return this.groupID;
    }

    public String getName() {
        return this.name;
    }

    public String getDescription() {
        return this.description;
    }

    public Iterator<Person> viewMembers() {
        return this.members.iterator();
    }

    public String setName(String name) {
        String oldName = this.name;
        this.name = name;
        return oldName;
    }

    public String setDescription(String description) {
        String oldDescription = this.description;
        this.description = description;
        return oldDescription;
    }

    public boolean addMember(Person person) {
        return this.members.add(person);
    }

    public boolean addMembers(Set<Person> people) {
        return this.members.addAll(people);
    }

    @Override
    public int compareTo(Group other) {
        return Integer.compare(this.groupID, other.groupID);
    }

    @Override
    public int hashCode() {
        return this.groupID;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) return true;
        if (!(other instanceof Group)) return false;

        return this.groupID == ((Group) other).groupID;
    }
}
